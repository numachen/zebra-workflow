# ZebraWorkflow

### 项目：基于 go-zero 的 HTTP + Temporal 工作流引擎，实现通过 HTTP 触发与管理工作流。

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
