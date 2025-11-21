package handler

import (
	"net/http"

	"zebra-workflow/internal/temporal"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterRoutes 注册 workflow 相关的路由
func RegisterRoutes(srv *rest.Server, tc *temporal.ClientWrapper) {
	// Start workflow
	srv.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/workflow/start",
		Handler: StartWorkflowHandler(tc),
	})

	// Query status
	srv.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/v1/workflow/:workflowId/status",
		Handler: QueryStatusHandler(tc),
	})

	// Signal
	srv.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/v1/workflow/:workflowId/signal",
		Handler: SignalHandler(tc),
	})

	// Info (optional)
	srv.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/v1/info",
		Handler: InfoHandler(),
	})

	// Swagger UI 页面（使用 CDN 的 Swagger UI，指定我们的 spec 地址）
	srv.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/swagger",
		Handler: SwaggerUIHandler(),
	})

	// Swagger OpenAPI 规范（静态文件），会读取仓库中的 docs/openapi.yaml
	srv.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/swagger/openapi.yaml",
		Handler: SwaggerSpecHandler(),
	})
}
