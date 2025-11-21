package workflow

import (
	"time"

	"zebra-workflow/internal/activity"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func init() {
	// 注册示例 workflow
	Register(&RegisteredWorkflow{
		Name:    "SampleWorkflow",
		Version: "v1",
		Factory: func() interface{} { return SampleWorkflow },
		Default: true,
	})
}

// SampleWorkflow 示例工作流（入口函数）
func SampleWorkflow(ctx workflow.Context, version string, input map[string]interface{}) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleWorkflow started", "version", version, "input", input)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 5,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	// 这里使用包级函数 activity.DoSomethingActivity（见 activity 包）
	err := workflow.ExecuteActivity(ctx, activity.DoSomethingActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("activity failed", "err", err)
		return err
	}

	// 工作流后续逻辑...
	logger.Info("SampleWorkflow finished", "result", result)
	return nil
}
