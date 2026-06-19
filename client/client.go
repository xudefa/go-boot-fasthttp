// Package client 基于 valyala/fasthttp 提供高性能 HTTP 客户端实现。
//
// 该包将 fasthttp 客户端与 go-boot 容器系统集成，
// 提供高性能的 HTTP 请求工具。
//
// 定义:
//
//   - HttpClient: HTTP 客户端实现了 net.HttpClient 接口
//   - HttpClientOption: 客户端配置选项
//
// 快速开始:
//
//	client, _ := client.NewHttpClient()
//	resp, _ := client.Get(context.Background(), "/api/hello")
//	fmt.Println(string(resp.Body))
package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xudefa/go-boot/net"

	"github.com/valyala/fasthttp"
)

// HttpClient 是 fasthttp HTTP 客户端，实现了 net.HttpClient 接口。
//
// 字段说明:
//   - client: fasthttp 客户端实例
//   - baseURL: 基础 URL
//   - timeout: 请求超时
//   - tlsConfig: TLS 配置
//   - defaultHeaders: 默认请求头，会应用到所有请求
type HttpClient struct {
	client         *fasthttp.Client
	baseURL        string
	timeout        time.Duration
	tlsConfig      *tls.Config
	defaultHeaders http.Header
}

// HttpClientOption 是客户端配置选项函数。
type HttpClientOption func(*HttpClient)

// NewHttpClient 创建新的 fasthttp HTTP 客户端。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *HttpClient: 配置好的客户端实例
//   - error: 创建错误
func NewHttpClient(opts ...HttpClientOption) (*HttpClient, error) {
	c := &HttpClient{
		client: &fasthttp.Client{
			MaxIdleConnDuration: 90 * time.Second,
			ReadTimeout:         30 * time.Second,
			WriteTimeout:        30 * time.Second,
			MaxConnsPerHost:     512,
		},
		baseURL:        "http://localhost:8080",
		timeout:        30 * time.Second,
		defaultHeaders: make(http.Header),
	}

	for _, opt := range opts {
		opt(c)
	}

	// 应用 TLS 配置
	if c.tlsConfig != nil {
		c.client.TLSConfig = c.tlsConfig
	}

	return c, nil
}

// WithBaseURL 设置客户端的基础 URL。
//
// 参数:
//   - baseURL: 基础 URL 地址
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithBaseURL(baseURL string) HttpClientOption {
	return func(c *HttpClient) {
		c.baseURL = baseURL
	}
}

// WithMaxConnsPerHost 设置每个主机的最大连接数。
//
// 参数:
//   - n: 最大连接数
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithMaxConnsPerHost(n int) HttpClientOption {
	return func(c *HttpClient) {
		c.client.MaxConnsPerHost = n
	}
}

// WithMaxIdleConnDuration 设置空闲连接的最长时间。
//
// 参数:
//   - duration: 空闲连接持续时间
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithMaxIdleConnDuration(duration time.Duration) HttpClientOption {
	return func(c *HttpClient) {
		c.client.MaxIdleConnDuration = duration
	}
}

// WithReadTimeout 设置读取超时时间。
//
// 参数:
//   - timeout: 读取超时时间
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithReadTimeout(timeout time.Duration) HttpClientOption {
	return func(c *HttpClient) {
		c.client.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时时间。
//
// 参数:
//   - timeout: 写入超时时间
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithWriteTimeout(timeout time.Duration) HttpClientOption {
	return func(c *HttpClient) {
		c.client.WriteTimeout = timeout
	}
}

// WithTimeout 设置客户端的默认超时时间。
//
// 参数:
//   - timeout: 超时时间
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithTimeout(timeout time.Duration) HttpClientOption {
	return func(c *HttpClient) {
		c.timeout = timeout
	}
}

// WithTLSConfig 设置 TLS 配置。
//
// 参数:
//   - tlsConfig: TLS 配置
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithTLSConfig(tlsConfig *tls.Config) HttpClientOption {
	return func(c *HttpClient) {
		c.tlsConfig = tlsConfig
	}
}

// WithHeader 设置默认请求头，会应用到所有请求。
//
// 参数:
//   - key: 请求头名称
//   - value: 请求头值
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithHeader(key, value string) HttpClientOption {
	return func(c *HttpClient) {
		c.defaultHeaders.Set(key, value)
	}
}

// WithHeaders 设置多个默认请求头。
//
// 参数:
//   - headers: 请求头 map
//
// 返回值:
//   - HttpClientOption: 客户端配置选项函数
func WithHeaders(headers http.Header) HttpClientOption {
	return func(c *HttpClient) {
		for key, values := range headers {
			for _, value := range values {
				c.defaultHeaders.Add(key, value)
			}
		}
	}
}

// Get 发送 GET 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Get(ctx context.Context, url string, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodGet, url, nil, opts...)
}

// Post 发送 POST 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - body: 请求体
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Post(ctx context.Context, url string, body any, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodPost, url, body, opts...)
}

// Put 发送 PUT 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - body: 请求体
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Put(ctx context.Context, url string, body any, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodPut, url, body, opts...)
}

// Delete 发送 DELETE 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Delete(ctx context.Context, url string, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodDelete, url, nil, opts...)
}

// Head 发送 HEAD 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Head(ctx context.Context, url string, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodHead, url, nil, opts...)
}

// Options 发送 OPTIONS 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Options(ctx context.Context, url string, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodOptions, url, nil, opts...)
}

// Patch 发送 PATCH 请求。
//
// 参数:
//   - ctx: 上下文
//   - url: 请求路径
//   - body: 请求体
//   - opts: 可选的请求选项
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Patch(ctx context.Context, url string, body any, opts ...net.RequestOption) (*net.HttpResponse, error) {
	return c.doRequest(ctx, fasthttp.MethodPatch, url, body, opts...)
}

// Do 发送自定义 HTTP 请求。
//
// 参数:
//   - ctx: 上下文
//   - req: 请求对象 (*fasthttp.Request)
//
// 返回值:
//   - *net.HttpResponse: 响应对象
//   - error: 请求错误
func (c *HttpClient) Do(ctx context.Context, req any) (*net.HttpResponse, error) {
	fhReq, ok := req.(*fasthttp.Request)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *fasthttp.Request")
	}

	fhResp := &fasthttp.Response{}
	err := c.client.Do(fhReq, fhResp)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return c.buildResponse(fhResp), nil
}

// Close 关闭客户端连接。
func (c *HttpClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

func (c *HttpClient) buildURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return c.baseURL + path
}

func (c *HttpClient) doRequest(ctx context.Context, method string, url string, body any, opts ...net.RequestOption) (*net.HttpResponse, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	cfg := &net.HttpRequest{}
	for _, opt := range opts {
		opt(cfg)
	}

	fullURL := c.buildURL(url)
	if len(cfg.Query) > 0 {
		fullURL += "?" + cfg.Query.Encode()
	}

	req.Header.SetMethod(method)
	req.SetRequestURI(fullURL)

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for key, values := range cfg.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}

	if cfg.BasicAuth.Username != "" || cfg.BasicAuth.Password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(cfg.BasicAuth.Username + ":" + cfg.BasicAuth.Password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	if body != nil {
		data, err := marshalBody(body)
		if err != nil {
			return nil, err
		}
		req.SetBody(data)
	}

	timeout := c.timeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	var err error
	if timeout > 0 {
		err = c.client.DoTimeout(req, resp, timeout)
	} else {
		err = c.client.Do(req, resp)
	}
	if err != nil {
		return nil, err
	}

	return c.buildResponse(resp), nil
}

func (c *HttpClient) buildResponse(resp *fasthttp.Response) *net.HttpResponse {
	header := make(http.Header)
	for key, value := range resp.Header.All() {
		header.Set(string(key), string(value))
	}

	// 复制 resp.Body，避免 ReleaseResponse 回收池缓冲区后返回悬空引用
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return &net.HttpResponse{
		StatusCode: resp.StatusCode(),
		Header:     header,
		Body:       body,
	}
}

func marshalBody(v any) ([]byte, error) {
	switch data := v.(type) {
	case []byte:
		return data, nil
	case string:
		return []byte(data), nil
	case io.Reader:
		return io.ReadAll(data)
	default:
		return json.Marshal(v)
	}
}
