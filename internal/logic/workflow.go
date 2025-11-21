package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/conf"

	"zebra-workflow/internal/temporal"
	"zebra-workflow/internal/types"
)

// StartWorkflowLogic 业务层：真正调用 temporal client 启动 workflow
func StartWorkflowLogic(ctx context.Context, tc *temporal.ClientWrapper, req *types.StartReq) (*types.StartResp, error) {
	wid, rid, err := tc.StartWorkflow(ctx, req.Name, req.Version, req.Input)
	if err != nil {
		return nil, err
	}
	return &types.StartResp{WorkflowID: wid, RunID: rid}, nil
}

// QueryStatusLogic 查询 workflow 状态（封装 temporal 调用）
func QueryStatusLogic(ctx context.Context, tc *temporal.ClientWrapper, workflowID string) (map[string]interface{}, error) {
	return tc.QueryWorkflowStatus(ctx, workflowID)
}

// SignalLogic 向 workflow 发送 signal
func SignalLogic(ctx context.Context, tc *temporal.ClientWrapper, workflowID string, req *types.SignalReq) error {
	return tc.SendSignal(ctx, workflowID, req.SignalName, req.Payload)
}

// InfoLogic 读取 configs/config.yaml 并返回 InfoResp
func InfoLogic() (*types.InfoResp, error) {
	var cfg struct {
		HTTP struct {
			Addr string `yaml:"addr" json:"addr"`
		} `yaml:"http" json:"http"`
		Temporal struct {
			HostPort         string `yaml:"hostPort" json:"hostPort"`
			Namespace        string `yaml:"namespace" json:"namespace"`
			DefaultTaskQueue string `yaml:"defaultTaskQueue" json:"defaultTaskQueue"`
		} `yaml:"temporal" json:"temporal"`
	}

	if err := conf.Load("configs/config.yaml", &cfg); err != nil {
		return nil, err
	}
	httpAddr := cfg.HTTP.Addr
	if httpAddr == "" {
		httpAddr = ":8888"
	}
	resp := &types.InfoResp{
		HTTPAddr: httpAddr,
		Temporal: map[string]string{
			"hostPort":         cfg.Temporal.HostPort,
			"namespace":        cfg.Temporal.Namespace,
			"defaultTaskQueue": cfg.Temporal.DefaultTaskQueue,
		},
	}
	return resp, nil
}
