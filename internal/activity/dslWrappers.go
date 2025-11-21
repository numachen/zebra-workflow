package activity

import (
	"context"
	"encoding/json"
	"fmt"

	dslpkg "zebra-workflow/internal/dsl"
)

// Wrapper functions calling the SampleActivities methods so the activity type name is the simple function name.
// These functions will be registered on the worker with these names.

func SampleActivity(ctx context.Context, input map[string]string) (map[string]interface{}, error) {
	return (&dslpkg.SampleActivities{}).SampleActivity(ctx, input)
}

func GetTitle(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return (&dslpkg.SampleActivities{}).GetTitle(ctx, input)
}

// SendEmailInput 示例强类型输入结构体（根据你的业务调整字段）
type SendEmailInput struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// 包级函数：接收 map[string]string（由 workflow 传入），函数内部按需要转换为 struct 或按名字读取值。
// 注册为活动时使用这个函数名（SampleActivitySendEmail）

func SampleActivitySendEmail(ctx context.Context, input map[string]string) (string, error) {
	// 直接按 key 读取（简单且高效）
	to := input["to"]
	subject := input["subject"]
	body := input["body"]

	// 业务处理示例
	fmt.Printf("SendEmail activity: to=%s subject=%s body=%s\n", to, subject, body)

	// 返回示例结果
	return "email_sent_to_" + to, nil
}

// 如果你希望活动接收具体 struct，可以在 wrapper 中把 map 转为 struct 并调用实现方法

func SampleActivitySendEmailTyped(ctx context.Context, input map[string]string) (map[string]interface{}, error) {
	var in SendEmailInput
	// 使用 json marshal/unmarshal 将 map -> struct（适合值都为字符串的 map）
	b, err := json.Marshal(input)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if err := json.Unmarshal(b, &in); err != nil {
		return map[string]interface{}{}, err
	}
	fmt.Printf("SendEmail activity: %+v\n", input)
	// 调用实际实现（可为方法或函数）
	return doSendEmail(ctx, in)
}

func doSendEmail(ctx context.Context, in SendEmailInput) (map[string]interface{}, error) {
	// 实际发送逻辑（示例）
	fmt.Printf("email_sent_to_: %+v\n", in.To, 89)
	return map[string]interface{}{"email_sent_to_": in.To}, nil
}
