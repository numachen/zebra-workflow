package activity

import (
	"context"
	"fmt"

	logger "zebra-workflow/internal/log"
)

// ActivityImpl 活动实现
type ActivityImpl struct{}

// NewActivityImpl returns a new ActivityImpl
func NewActivityImpl() *ActivityImpl {
	return &ActivityImpl{}
}

// DoSomethingActivity 方法实现（receiver 方法，供 worker 注册实例时使用）
func (a *ActivityImpl) DoSomethingActivity(ctx context.Context, input map[string]interface{}) (string, error) {
	// 在这里实现业务逻辑（注意：activity 应保持幂等或在上层处理重试）
	// log, metrics, error handling...
	fmt.Println("activity received input:", input)
	logger.Sugar.Infow("activity.DoSomethingActivity finished", "result", input)
	// 模拟工作
	return "ok", nil
}

// DoSomethingActivity 包级函数封装，方便在 workflow.ExecuteActivity 中直接引用
// 它会创建一个临时实现实例并调用其方法。
// 也可以改为直接注册该函数到 worker：w.RegisterActivity(activity.DoSomethingActivity)
func DoSomethingActivity(ctx context.Context, input map[string]interface{}) (string, error) {
	return NewActivityImpl().DoSomethingActivity(ctx, input)
}
