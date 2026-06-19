// Package fasthttp 提供 Fasthttp HTTP 客户端的自动配置。
//
// 当 fasthttp.enabled=true 时自动启用，从 Environment 中读取 fasthttp.base-url、fasthttp.timeout、
// fasthttp.max-conns-per-host、fasthttp.read-timeout、fasthttp.write-timeout 等配置项，
// 创建并注册 Fasthttp HttpClient Bean 到 IoC 容器中（Bean ID: fastHttpClient），实现 net.HttpClient 接口。
package fasthttp

import (
	"time"

	"github.com/xudefa/go-boot-fasthttp/client"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/net"
)

// init 注册 Fasthttp 自动配置，由 fasthttp.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&FasthttpAutoConfiguration{},
		condition.OnProperty(constants.FastHTTPEnabled, constants.ConditionTrue),
	)
}

// FasthttpAutoConfiguration Fasthttp HTTP 客户端的自动配置。
//
// 从 Environment 中读取 fasthttp.base-url、fasthttp.timeout、fasthttp.max-conns-per-host 等配置项，
// 创建 Fasthttp HttpClient 实例并注册到 IoC 容器中，实现 net.HttpClient 接口。
// 启用条件：fasthttp.enabled=true
type FasthttpAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 Fasthttp HttpClient 并注册为 Bean。
func (f *FasthttpAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	opts := []client.HttpClientOption{}

	if baseURL := env.GetString(constants.FastHTTPBaseURL, ""); baseURL != "" {
		opts = append(opts, client.WithBaseURL(baseURL))
	}
	opts = append(opts,
		client.WithTimeout(time.Duration(env.GetInt(constants.FastHTTPTimeout, constants.DefaultFastHTTPTimeout))*time.Second),
		client.WithMaxConnsPerHost(env.GetInt(constants.FastHTTPMaxConnsPerHost, constants.DefaultFastHTTPMaxConnsPerHost)),
		client.WithReadTimeout(time.Duration(env.GetInt(constants.FastHTTPReadTimeout, constants.DefaultFastHTTPReadTimeout))*time.Second),
		client.WithWriteTimeout(time.Duration(env.GetInt(constants.FastHTTPWriteTimeout, constants.DefaultFastHTTPWriteTimeout))*time.Second),
	)

	httpClient, err := client.NewHttpClient(opts...)
	if err != nil {
		return err
	}

	if err := ctx.Register(constants.FastHTTPClientBeanID,
		core.Bean(httpClient),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}

var _ net.HttpClient = (*client.HttpClient)(nil)
