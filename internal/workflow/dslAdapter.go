package workflow

import (
	"encoding/json"
	"errors"

	dslpkg "zebra-workflow/internal/dsl"
	"zebra-workflow/internal/log"

	"go.temporal.io/sdk/workflow"
)

// DSLWorkflowWrapper 拆装参数并调用 samples 的 SimpleDSLWorkflow。
// 支持多种输入形态：
// 1. 直接传入 dsl.Workflow 的 JSON（推荐）
// 2. 将 DSL 包在 input["input"]（或 "Input"）里（兼容某些前端）
func init() {
	Register(&RegisteredWorkflow{
		Name:    "DSLWorkflow",
		Version: "v1",
		Factory: func() interface{} { return DSLWorkflowWrapper },
		Default: true,
	})
}

func DSLWorkflowWrapper(ctx workflow.Context, version string, input map[string]interface{}) error {
	// 尝试直接把 input 解析为 dsl.Workflow（最常见）
	var dslWorkflow dslpkg.Workflow
	b, err := json.Marshal(input)
	if err != nil {
		log.Sugar.Errorw("dsl adapter: marshal input failed", "err", err)
		return err
	}
	if err := json.Unmarshal(b, &dslWorkflow); err == nil {
		// 判断是否解析出至少一部分有意义的数据（Root 不为空或 Variables 不空）
		if dslWorkflow.Root.Activity != nil || dslWorkflow.Root.Sequence != nil || dslWorkflow.Root.Parallel != nil || len(dslWorkflow.Variables) > 0 {
			log.Sugar.Infow("dsl adapter: input parsed as direct Workflow", "version", version)
			_, err = dslpkg.SimpleDSLWorkflow(ctx, dslWorkflow)
			return err
		}
	}

	// 若直接解析后内容为空，则尝试解包常见的嵌套键 "input" / "Input"
	if nestedRaw, ok := input["input"]; ok {
		if nestedMap, ok2 := nestedRaw.(map[string]interface{}); ok2 {
			// marshal nested and unmarshal
			b2, merr := json.Marshal(nestedMap)
			if merr != nil {
				log.Sugar.Errorw("dsl adapter: marshal nested input failed", "err", merr)
				return merr
			}
			if err := json.Unmarshal(b2, &dslWorkflow); err != nil {
				log.Sugar.Errorw("dsl adapter: unmarshal nested input failed", "err", err)
				return err
			}
			log.Sugar.Infow("dsl adapter: nested input parsed and will be executed", "version", version)
			_, err = dslpkg.SimpleDSLWorkflow(ctx, dslWorkflow)
			return err
		}
	}

	// 也尝试 "Input"（大写）键以兼容不同客户端
	if nestedRaw, ok := input["Input"]; ok {
		if nestedMap, ok2 := nestedRaw.(map[string]interface{}); ok2 {
			b2, merr := json.Marshal(nestedMap)
			if merr != nil {
				log.Sugar.Errorw("dsl adapter: marshal Input failed", "err", merr)
				return merr
			}
			if err := json.Unmarshal(b2, &dslWorkflow); err != nil {
				log.Sugar.Errorw("dsl adapter: unmarshal Input failed", "err", err)
				return err
			}
			log.Sugar.Infow("dsl adapter: Input parsed and will be executed", "version", version)
			_, err = dslpkg.SimpleDSLWorkflow(ctx, dslWorkflow)
			return err
		}
	}

	// 如果都不能解析出合理 DSL，返回错误提示
	log.Sugar.Errorw("dsl adapter: unable to parse workflow input into DSL", "inputKeys", keysOf(input))
	return errors.New("invalid dsl input: expected dsl.Workflow structure or nested 'input' object")
}

// keysOf 辅助函数：返回 map 的 key 列表（用于日志）
func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
