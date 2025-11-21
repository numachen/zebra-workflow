package workflow

import "go.temporal.io/sdk/workflow"

// 定义 workflow 注册信息

type WorkflowFactory func() interface{}

type RegisteredWorkflow struct {
	Name    string
	Version string
	Factory WorkflowFactory
	Default bool
}

var registry = make([]*RegisteredWorkflow, 0)

func Register(w *RegisteredWorkflow) {
	registry = append(registry, w)
}

func ListRegistered() []*RegisteredWorkflow {
	return registry
}

func GetRegisterOptions(r *RegisteredWorkflow) workflow.RegisterOptions {
	// 可以根据 version 设置不同的 options，比方说 name 带 version 后缀
	return workflow.RegisterOptions{Name: r.Name}
}
