// Package middleware 提供 Fasthttp 追踪中间件。
//
// 包含服务器端和客户端追踪中间件，支持上下文传播和 TraceID 注入。
package middleware

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/xudefa/go-boot/tracing"
)

// fhCarrier 实现 tracing.TextMapCarrier 接口，用于在 Fasthttp 请求中提取和注入追踪上下文
type fhCarrier struct {
	ctx *fasthttp.RequestCtx
}

// Get 获取指定键的 header 值
func (c *fhCarrier) Get(key string) string {
	return string(c.ctx.Request.Header.Peek(key))
}

// Set 设置指定键的 header 值
func (c *fhCarrier) Set(key string, value string) {
	c.ctx.Request.Header.Set(key, value)
}

// Keys 返回所有 header 键的列表
func (c *fhCarrier) Keys() []string {
	keys := make([]string, 0)
	for key := range c.ctx.Request.Header.All() {
		keys = append(keys, string(key))
	}
	return keys
}

// GetTraceID 从 context.Context 中提取当前 Span 的 TraceID
// 返回空字符串如果没有有效的追踪上下文
func GetTraceID(ctx context.Context) string {
	span := tracing.SpanFromContext(ctx)
	if span == nil || span.GetTraceID() == "" {
		return ""
	}
	return span.GetTraceID()
}

// GetSpanID 从 context.Context 中提取当前 Span 的 SpanID
// 返回空字符串如果没有有效的追踪上下文
func GetSpanID(ctx context.Context) string {
	span := tracing.SpanFromContext(ctx)
	if span == nil || span.GetSpanID() == "" {
		return ""
	}
	return span.GetSpanID()
}

// AddTraceToResponseHeaders 将 TraceID 和 SpanID 添加到响应头中
// 便于客户端获取追踪信息进行问题排查
func AddTraceToResponseHeaders(ctx context.Context, fctx *fasthttp.RequestCtx) {
	traceID := GetTraceID(ctx)
	spanID := GetSpanID(ctx)
	if traceID != "" {
		fctx.Response.Header.Set("X-Trace-ID", traceID)
	}
	if spanID != "" {
		fctx.Response.Header.Set("X-Span-ID", spanID)
	}
}

// TraceIDMiddleware 简单的中间件，仅将追踪 ID 添加到响应头
// 适用于不需要完整追踪功能但需要暴露追踪 ID 的场景
func TraceIDMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		traceCtx := ctx.UserValue("traceContext")
		if tc, ok := traceCtx.(context.Context); ok {
			AddTraceToResponseHeaders(tc, ctx)
		}
		if next != nil {
			next(ctx)
		}
	}
}

// HTTPServerTracingMiddleware 创建 Fasthttp HTTP 服务器端追踪中间件
// serviceName 参数用于标识服务名称，默认为 "fasthttp-server"
//
// 该中间件提供以下功能：
// 1. 从请求头中提取父追踪上下文
// 2. 创建服务端 Span，记录 HTTP 方法、路径、主机等信息
// 3. 将 Span 上下文存储到 RequestCtx 的 UserValue 中
// 4. 在请求结束时记录响应状态码和错误状态
// 5. 将 TraceID/SpanID 添加到响应头
func HTTPServerTracingMiddleware(serviceName ...string) func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	tracerName := "fasthttp-server"
	if len(serviceName) > 0 {
		tracerName = serviceName[0]
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			carrier := &fhCarrier{ctx: ctx}
			parentCtx := tracing.ExtractTraceContext(context.Background(), carrier)

			spanName := string(ctx.Request.RequestURI())
			if spanName == "" {
				spanName = "HTTP " + string(ctx.Method())
			}

			tracer := tracing.GetTracer(tracerName)
			ctx2, span := tracing.StartHTTPServerSpan(parentCtx, tracer, spanName,
				string(ctx.Method()),
				string(ctx.Request.RequestURI()),
				string(ctx.Host()),
			)
			defer span.End()

			ctx.SetUserValue("traceContext", ctx2)

			if next != nil {
				next(ctx)
			}

			statusCode := ctx.Response.StatusCode()
			span.SetAttribute("http.status_code", statusCode)

			if statusCode >= 500 {
				span.SetStatus(tracing.SpanStatusError)
			} else {
				span.SetStatus(tracing.SpanStatusOK)
			}

			AddTraceToResponseHeaders(ctx2, ctx)
		}
	}
}

// InjectTraceHeaders 将当前追踪上下文注入到请求头中
// 用于客户端向外发起请求时传递追踪信息
func InjectTraceHeaders(ctx context.Context, fctx *fasthttp.RequestCtx) {
	tracing.InjectTraceContext(ctx, &fhCarrier{ctx: fctx})
}

// HTTPClientTracingMiddleware 创建 Fasthttp HTTP 客户端追踪中间件
// 用于在向外发起 HTTP 请求时注入追踪上下文
func HTTPClientTracingMiddleware(serviceName ...string) func(ctx context.Context, fctx *fasthttp.RequestCtx) {
	return func(ctx context.Context, fctx *fasthttp.RequestCtx) {
		InjectTraceHeaders(ctx, fctx)
	}
}

// NewTracedClient 创建一个带有追踪功能的 Fasthttp 客户端
func NewTracedClient() *fasthttp.Client {
	c := &fasthttp.Client{
		MaxIdleConnDuration: 90 * time.Second,
	}
	return c
}

// 编译期类型断言，确保 fhCarrier 实现了 TextMapCarrier 接口
var _ tracing.TextMapCarrier = (*fhCarrier)(nil)
