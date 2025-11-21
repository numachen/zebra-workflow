package handler

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"zebra-workflow/internal/log"
	"zebra-workflow/internal/logic"
	"zebra-workflow/internal/temporal"
	"zebra-workflow/internal/types"
)

// StartWorkflowHandler HTTP 层：解析请求 -> 调用 logic -> 返回
func StartWorkflowHandler(tc *temporal.ClientWrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StartReq
		if err := httpx.Parse(r, &req); err != nil {
			log.Sugar.Warnw("parse start request failed", "error", err)
			httpx.Error(w, err)
			return
		}

		log.Sugar.Infow("start workflow request", "name", req.Name, "version", req.Version)
		resp, err := logic.StartWorkflowLogic(r.Context(), tc, &req)
		if err != nil {
			log.Sugar.Errorw("start workflow failed", "error", err, "name", req.Name)
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// QueryStatusHandler HTTP 层：解析 path -> 调用 logic -> 返回
func QueryStatusHandler(tc *temporal.ClientWrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wid := extractWorkflowID(r)
		if wid == "" {
			log.Sugar.Warn("query status failed: workflowId not found in path")
			http.Error(w, "workflowId not found in path", http.StatusBadRequest)
			return
		}
		log.Sugar.Infow("query workflow status", "workflowId", wid)
		resp, err := logic.QueryStatusLogic(r.Context(), tc, wid)
		if err != nil {
			log.Sugar.Errorw("query workflow status failed", "workflowId", wid, "error", err)
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SignalHandler HTTP 层：解析 body + path -> 调用 logic -> 返回
func SignalHandler(tc *temporal.ClientWrapper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SignalReq
		if err := httpx.Parse(r, &req); err != nil {
			log.Sugar.Warnw("parse signal request failed", "error", err)
			httpx.Error(w, err)
			return
		}
		wid := extractWorkflowID(r)
		if wid == "" {
			log.Sugar.Warn("signal request missing workflowId")
			http.Error(w, "workflowId not found in path", http.StatusBadRequest)
			return
		}
		log.Sugar.Infow("sending signal to workflow", "workflowId", wid, "signal", req.SignalName)
		if err := logic.SignalLogic(r.Context(), tc, wid, &req); err != nil {
			log.Sugar.Errorw("send signal failed", "workflowId", wid, "error", err)
			httpx.Error(w, err)
			return
		}
		httpx.Ok(w)
	}
}

// InfoHandler 仍然保留（读取 configs/config.yaml 并返回，具体实现可以复用已有代码）
func InfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 复用此前实现（或从 logic 中抽取），这里为了简洁直接重用旧实现
		infoResp, err := logic.InfoLogic()
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, infoResp)
	}
}

// SwaggerUIHandler 返回一个使用 Swagger UI CDN 的页面，页面会加载 /swagger/openapi.yaml
func SwaggerUIHandler() http.HandlerFunc {
	const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.css" integrity="" crossorigin="anonymous"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-bundle.min.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: '/swagger/openapi.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis
        ],
        layout: "BaseLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}

// SwaggerSpecHandler 直接返回仓库内的 docs/openapi.yaml（作为 swagger 的 spec）
func SwaggerSpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		specPath := "docs/openapi.yaml"
		data, err := ioutil.ReadFile(specPath)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

// extractWorkflowID 从 URL path 中按约定提取 /v1/workflow/<id>/... 中的 id
func extractWorkflowID(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if p == "workflow" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
