# Monitoring & Logging

建议实现:
- Prometheus metrics:
    - HTTP 服务: 请求总量、请求耗时、错误率
    - Worker/Activity: 执行次数、失败次数、重试次数、处理时延
    - Job queue depth（如 Temporal task queue backlog）
- Logs:
    - 使用 zap 或 go-zero 默认日志组件，输出 structured logs（time, level, workflowId, runId, traceId）
    - 将日志推到集中式系统（Loki / ELK）
- Alerting:
    - 配置错误率阈值、重试失败次数阈值、worker 掉线告警
- Tracing:
    - 推荐使用 OpenTelemetry，跨请求 / workflow 链路追踪（通过上下文传播 traceId）

Prometheus 示例:
- expose /metrics endpoint on :9090
- 使用 prometheus client_golang 在代码中声明指标