# go-boot-fasthttp

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-fasthttp)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-fasthttp)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-fasthttp/test.yml?branch=master)](https://github.com/xudefa/go-boot-fasthttp/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-fasthttp.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-fasthttp) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-fasthttp)](https://goreportcard.com/report/github.com/xudefa/go-boot-fasthttp)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 FastHTTP 高性能 HTTP 客户端集成模块。将 FastHTTP 无缝集成到 go-boot 的 IoC 容器和自动配置体系中,提供高性能的 HTTP 请求能力。

> 设计理念:遵循 go-boot 的开发规范,将 FastHTTP Client 作为 `net.HttpClient` 接口的实现,通过自动配置实现零代码启动 HTTP 客户端服务。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   go-boot-fasthttp Starter    │
                    │  ┌─────────────────────────┐  │
                    │  │ FastHttpClient Bean     │  │
                    │  │ (net.HttpClient)        │  │
                    │  │ Connection Pool         │  │
                    │  │ Request/Response        │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [HTTP 请求](#http-请求)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 FastHTTP 集成模块
go get github.com/xudefa/go-boot-fasthttp
```

### 最小示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/xudefa/go-boot/boot"
    fasthttp "github.com/xudefa/go-boot-fasthttp"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-http-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProperty("fasthttp.enabled", "true"),
        boot.WithProperty("fasthttp.base-url", "http://localhost:8080"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 启动应用（自动创建 FastHTTP Client）
    app.Start()

    // 从容器获取 HTTP 客户端
    client := app.Container().Get("fastHttpClient").(*fasthttp.HttpClient)

    // 发送 HTTP 请求
    resp, err := client.Get(context.Background(), "/api/hello")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(resp.Body))

    // 等待终止信号
    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| FastHTTP 集成 | 将 FastHTTP Client 注册为 Bean,支持依赖注入 |
| net.HttpClient 实现 | 实现 go-boot 的 `net.HttpClient` 接口 |
| 自动配置 | 通过 `fasthttp.enabled=true` 自动创建 HTTP 客户端 |
| 高性能 | 基于 FastHTTP 的高性能 HTTP 请求能力 |
| 连接池管理 | 内置连接池,自动管理空闲连接 |
| 配置驱动 | 超时、连接数、TLS 等均可通过配置控制 |
| 请求选项 | 支持自定义请求头、认证、超时等选项 |

## HTTP 请求

### 基本请求

```go
client, _ := fasthttp.NewHttpClient(
    fasthttp.WithBaseURL("http://api.example.com"),
    fasthttp.WithTimeout(30*time.Second),
)

// GET 请求
resp, err := client.Get(context.Background(), "/api/users")

// POST 请求
resp, err := client.Post(context.Background(), "/api/users", map[string]string{
    "name": "John",
    "email": "john@example.com",
})

// PUT 请求
resp, err := client.Put(context.Background(), "/api/users/1", map[string]string{
    "name": "John Updated",
})

// DELETE 请求
resp, err := client.Delete(context.Background(), "/api/users/1")
```

### 带请求选项

```go
import "github.com/xudefa/go-boot/net"

// 带认证和自定义请求头
resp, err := client.Get(context.Background(), "/api/secure",
    net.WithAuthToken("your-token"),
    net.WithHeader("X-Custom-Header", "custom-value"),
    net.WithTimeout(10*time.Second),
)

// 带 Basic Auth
resp, err := client.Get(context.Background(), "/api/auth",
    net.WithBasicAuth("username", "password"),
)

// 带查询参数
resp, err := client.Get(context.Background(), "/api/search",
    net.WithQuery("page", "1"),
    net.WithQuery("size", "20"),
)
```

### TLS 配置

```go
import "crypto/tls"

client, _ := fasthttp.NewHttpClient(
    fasthttp.WithTLSConfig(&tls.Config{
        InsecureSkipVerify: false,
    }),
)
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置:

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `fasthttp.enabled` | `false` | 是否启用 FastHTTP 客户端 |
| `fasthttp.base-url` | `http://localhost:8080` | 基础 URL 地址 |
| `fasthttp.timeout` | `30` | 请求超时(秒) |
| `fasthttp.max-conns-per-host` | `512` | 每个主机的最大连接数 |
| `fasthttp.read-timeout` | `30` | 读取超时(秒) |
| `fasthttp.write-timeout` | `30` | 写入超时(秒) |

### 示例配置

```yaml
# application.yml
fasthttp:
  enabled: true
  base-url: http://api.example.com
  timeout: 30
  max-conns-per-host: 512
  read-timeout: 30
  write-timeout: 30
```

## 项目结构

```
go-boot-fasthttp/
├── fasthttp.go             # FastHTTP Client 核心实现
├── autoconfig.go           # 自动配置注册
├── middleware_tracing.go   # 分布式追踪中间件
├── fasthttp_test.go        # 单元测试
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request!详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。