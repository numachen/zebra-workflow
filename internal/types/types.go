package types

// StartReq 请求体: 启动 workflow
type StartReq struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
}

// StartResp 启动 workflow 后返回
type StartResp struct {
	WorkflowID string `json:"workflowId"`
	RunID      string `json:"runId"`
}

// SignalReq 发送 signal 的请求体
type SignalReq struct {
	SignalName string                 `json:"signalName"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
}

// InfoResp / Query 接口的简单响应（可按需扩展）
type InfoResp struct {
	HTTPAddr string            `json:"httpAddr"`
	Temporal map[string]string `json:"temporal"`
}

type ArticleInfo struct {
	Title   string `json:"title"`
	Time    string `json:"time"`
	Content string `json:"content"`
}
