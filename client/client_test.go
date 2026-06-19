package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xudefa/go-boot/net"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpClient(t *testing.T) {
	client, err := NewHttpClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.Equal(t, 30*time.Second, client.timeout)
}

func TestNewHttpClientWithOptions(t *testing.T) {
	client, err := NewHttpClient(
		WithBaseURL("https://api.example.com"),
		WithTimeout(10*time.Second),
		WithMaxConnsPerHost(100),
		WithMaxIdleConnDuration(5*time.Minute),
		WithReadTimeout(5*time.Second),
		WithWriteTimeout(5*time.Second),
		WithHeader("User-Agent", "TestAgent"),
		WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://api.example.com", client.baseURL)
	assert.Equal(t, 10*time.Second, client.timeout)
	assert.Equal(t, 100, client.client.MaxConnsPerHost)
	assert.Equal(t, 5*time.Minute, client.client.MaxIdleConnDuration)
	assert.Equal(t, 5*time.Second, client.client.ReadTimeout)
	assert.Equal(t, 5*time.Second, client.client.WriteTimeout)
	assert.Equal(t, "TestAgent", client.defaultHeaders.Get("User-Agent"))
}

func TestWithBaseURL(t *testing.T) {
	client, err := NewHttpClient(WithBaseURL("https://example.com"))
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", client.baseURL)
}

func TestWithTimeout(t *testing.T) {
	client, err := NewHttpClient(WithTimeout(5 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, client.timeout)
}

func TestWithMaxConnsPerHost(t *testing.T) {
	client, err := NewHttpClient(WithMaxConnsPerHost(200))
	assert.NoError(t, err)
	assert.Equal(t, 200, client.client.MaxConnsPerHost)
}

func TestWithMaxIdleConnDuration(t *testing.T) {
	client, err := NewHttpClient(WithMaxIdleConnDuration(10 * time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Minute, client.client.MaxIdleConnDuration)
}

func TestWithReadTimeout(t *testing.T) {
	client, err := NewHttpClient(WithReadTimeout(15 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 15*time.Second, client.client.ReadTimeout)
}

func TestWithWriteTimeout(t *testing.T) {
	client, err := NewHttpClient(WithWriteTimeout(15 * time.Second))
	assert.NoError(t, err)
	assert.Equal(t, 15*time.Second, client.client.WriteTimeout)
}

func TestWithTLSConfig(t *testing.T) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	client, err := NewHttpClient(WithTLSConfig(tlsConfig))
	assert.NoError(t, err)
	assert.Equal(t, tlsConfig, client.tlsConfig)
	assert.Equal(t, tlsConfig, client.client.TLSConfig)
}

func TestWithHeader(t *testing.T) {
	client, err := NewHttpClient(WithHeader("Authorization", "Bearer token"))
	assert.NoError(t, err)
	assert.Equal(t, "Bearer token", client.defaultHeaders.Get("Authorization"))
}

func TestWithHeaders(t *testing.T) {
	headers := make(http.Header)
	headers.Set("X-Test-Header", "test-value")
	headers.Set("X-Another-Header", "another-value")

	client, err := NewHttpClient(WithHeaders(headers))
	assert.NoError(t, err)
	assert.Equal(t, "test-value", client.defaultHeaders.Get("X-Test-Header"))
	assert.Equal(t, "another-value", client.defaultHeaders.Get("X-Another-Header"))
}

func TestBuildURL(t *testing.T) {
	client, _ := NewHttpClient(WithBaseURL("https://api.example.com"))

	url := client.buildURL("/users")
	assert.Equal(t, "https://api.example.com/users", url)

	url = client.buildURL("http://other.com/data")
	assert.Equal(t, "http://other.com/data", url)

	url = client.buildURL("https://secure.com/data")
	assert.Equal(t, "https://secure.com/data", url)
}

func TestMarshalBody(t *testing.T) {
	data := []byte("test data")
	result, err := marshalBody(data)
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	str := "test string"
	result, err = marshalBody(str)
	assert.NoError(t, err)
	assert.Equal(t, []byte(str), result)

	obj := struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}{
		Name: "test",
		ID:   123,
	}
	result, err = marshalBody(obj)
	assert.NoError(t, err)
	expected := `{"name":"test","id":123}`
	assert.Equal(t, expected, string(result))
}

func TestClose(t *testing.T) {
	client, err := NewHttpClient()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestHttpClientIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Hello, World!"}`))
	}))
	defer server.Close()

	client, err := NewHttpClient(WithBaseURL(server.URL))
	assert.NoError(t, err)

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(resp.Body), "Hello, World!")
}

func TestHttpClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received": true}`))
	}))
	defer server.Close()

	client, err := NewHttpClient(WithBaseURL(server.URL))
	assert.NoError(t, err)

	ctx := context.Background()
	payload := map[string]any{
		"name":  "test",
		"value": 123,
	}
	resp, err := client.Post(ctx, "/post", payload)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHttpClientDo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("custom response"))
	}))
	defer server.Close()

	client, err := NewHttpClient(WithBaseURL(server.URL))
	assert.NoError(t, err)

	_, err = client.Do(context.Background(), "invalid type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid request type")
}

var _ net.HttpClient = (*HttpClient)(nil)
