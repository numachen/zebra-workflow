package temporal

import (
	"context"
	"fmt"
	"time"
	logger "zebra-workflow/internal/log"

	"go.temporal.io/sdk/client"

	"github.com/zeromicro/go-zero/core/conf"
)

// Config 映射 configs/config.yaml 中需要的 temporal 字段
type Config struct {
	Temporal struct {
		HostPort         string `json:"hostPort" yaml:"hostPort"`
		Namespace        string `json:"namespace" yaml:"namespace"`
		DefaultTaskQueue string `json:"defaultTaskQueue" yaml:"defaultTaskQueue"`
	} `json:"temporal" yaml:"temporal"`
}

// ClientWrapper 包装 Temporal client 提供封装方法（start/query/signal）
type ClientWrapper struct {
	cli          client.Client
	namespace    string
	defaultQueue string
}

// NewClientFromConfig 读取配置并初始化 client
func NewClientFromConfig(configPath string) (*ClientWrapper, error) {
	var cfg Config
	// 使用 go-zero 的 conf.Load 读取配置文件到 cfg
	if err := conf.Load(configPath, &cfg); err != nil {
		logger.Sugar.Warnw("failed to load temporal config, falling back to defaults", "err", err)
		return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	// 从配置读取，若为空则使用默认值
	hostPort := cfg.Temporal.HostPort
	if hostPort == "" {
		hostPort = "127.0.0.1:7233"
	}
	namespace := cfg.Temporal.Namespace
	if namespace == "" {
		namespace = "default"
	}
	defaultQueue := cfg.Temporal.DefaultTaskQueue
	if defaultQueue == "" {
		defaultQueue = "zebra-task-queue"
	}

	cli, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		logger.Sugar.Errorw("unable to create temporal client", "err", err)
		return nil, fmt.Errorf("unable to create temporal client: %w", err)
	}
	return &ClientWrapper{cli: cli, namespace: namespace, defaultQueue: defaultQueue}, nil
}

func (c *ClientWrapper) Close() {
	c.cli.Close()
}

// Client 返回底层 temporal.Client，供外部（worker）使用
func (c *ClientWrapper) Client() client.Client {
	return c.cli
}

// DefaultQueue 返回默认 task queue 名称
func (c *ClientWrapper) DefaultQueue() string {
	return c.defaultQueue
}

// StartWorkflow 启动 workflow，name + version 决定其实例化工厂（registry 中的工厂）
func (c *ClientWrapper) StartWorkflow(ctx context.Context, name string, version string, input interface{}) (workflowID string, runID string, err error) {
	// workflowID 可自定义，或直接使用 temporal 生成的
	workflowID = fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: c.defaultQueue,
	}
	// 可扩展：根据 name/version 设置不同的 retry / timeouts / memo 等
	we, err := c.cli.ExecuteWorkflow(ctx, options, name, version, input)
	if err != nil {
		return "", "", err
	}
	return we.GetID(), we.GetRunID(), nil
}

// QueryWorkflowStatus 查询当前 workflow 的状态（简单返回基本信息）
func (c *ClientWrapper) QueryWorkflowStatus(ctx context.Context, workflowID string) (map[string]interface{}, error) {
	// 此处我们演示使用 DescribeWorkflowExecution
	desc, err := c.cli.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		logger.Sugar.Errorw("describe workflow execution failed", "workflowId", workflowID, "err", err)
		return nil, err
	}
	resp := map[string]interface{}{
		"workflowExecutionInfo": desc,
	}
	return resp, nil
}

// SendSignal 向运行中的 workflow 发送 signal
func (c *ClientWrapper) SendSignal(ctx context.Context, workflowID string, signalName string, payload interface{}) error {
	if err := c.cli.SignalWorkflow(ctx, workflowID, "", signalName, payload); err != nil {
		logger.Sugar.Errorw("signal workflow failed", "workflowId", workflowID, "err", err)
		return err
	}
	return nil
}
