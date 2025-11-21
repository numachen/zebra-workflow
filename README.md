# ZebraWorkflow

ZebraWorkflow 是一个基于 go-zero 框架的 HTTP + Temporal 工作流引擎项目。该项目通过 HTTP 接口触发和管理工作流执行，支持 DSL（领域特定语言）定义工作流逻辑。项目提供了完整的工作流编排能力，包括活动（Activity）的注册、顺序执行、参数传递和结果返回。支持多种部署方式，包括本地开发和 Docker Compose 容器化部署，并提供 Swagger 文档和日志记录功能。

### 主要功能点
- 基于 HTTP 的工作流启动和管理接口
- DSL 工作流定义和执行引擎
- Temporal 分布式工作流框架集成
- 活动（Activity）注册和管理机制
- 顺序执行（Sequence）的工作流编排
- Swagger API 文档支持
- 日志记录和追踪功能
- 多部署方式支持（本地开发、Docker Compose）

### 技术栈
- 语言：Go 100%
- 核心框架：go-zero（微服务框架）、Temporal（分布式工作流）
- 日志库：Zap（uber-go/zap）
- 配置管理：fsnotify（文件系统监听）
- 容器化：Docker Compose
- API 文档：Swagger

### 项目初始化
```shell
go mod init zebra-workflow
go get -u github.com/zeromicro/go-zero@latest

# 安装 temporal
go get -u go.temporal.io/sdk
go get -u go.temporal.io/sdk/client

# 安装 zap
go get -u go.uber.org/zap
go get -u github.com/fsnotify/fsnotify
go mod tidy
```

### 运行
```shell
# 启动worker
go run .\cmd\worker\main.go
# 启动server
go run .\cmd\server\main.go
```

### 功能
- Swagger地址：http://localhost:8080/swagger
- 日志记录

### 服务端部署方式
1. 方式一：[本地开发方式](https://docs.temporal.io/develop/go/set-up-your-local-go)
2. 方式二：docker-compose方式

### 开发流程
1. 请求示例
```curl
curl -X 'POST' \
  'http://127.0.0.1:8888/v1/workflow/start' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "DSLWorkflow",
    "version": "v1",
    "input": {
      "Variables": {
        "to": "a@example.com",
        "subject": "hi",
        "body": "hello from DSL"
      },
      "Root": {
        "Sequence": {
          "Elements": [
            { "Activity": { "Name": "SampleActivity", "Arguments": ["to","subject","body"], "Result": "r1" } },
            { "Activity": { "Name": "GetTitle", "Arguments": [ "r1"], "Result": "article" }}
          ]
        }
      }
    }
  }'
 ```

2. 给某人发邮件
```curl
curl -X 'POST' \
  'http://127.0.0.1:8888/v1/workflow/start' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "DSLWorkflow",
    "version": "v1",
    "input": {
      "Variables": {
        "to": "a@example.com",
        "subject": "hi",
        "body": "hello from DSL"
      },
      "Root": {
        "Sequence": {
          "Elements": [
            { "Activity": { "Name": "SampleActivitySendEmailTyped", "Arguments": ["to","subject","body"], "Result": "r1" } }
          ]
        }
      }
    }
  }'
```


## 开发笔记
1. 项目中有两个工作流：DSLWorkflow、SampleWorkflow；
2. 统一在workflow/registry.go中注册；
