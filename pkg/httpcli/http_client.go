package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	Timeout             int64 `yaml:"timeout,omitempty"`                 // 总超时时间（秒，推荐 5-10s）
	MaxIdleConns        int   `yaml:"max_idle_conns,omitempty"`          // 最大空闲连接数
	MaxIdleConnsPerHost int   `yaml:"max_idle_conns_per_host,omitempty"` // 每个Host最大空闲连接
	IdleConnTimeout     int64 `yaml:"idle_conn_timeout,omitempty"`       // 空闲连接超时（秒）
}

func defaultConfig() *Config {
	return &Config{
		Timeout:             5, // * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30, // * time.Second,
	}
}

type RequestOption func(r *http.Request)

func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

func WithContext(ctx context.Context) RequestOption {
	return func(r *http.Request) {
		*r = *r.WithContext(ctx)
	}
}

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{}
}

func (c *HttpClient) Init(cfg *Config) error {
	if cfg == nil {
		cfg = defaultConfig()
	}
	c.client = &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Duration(cfg.IdleConnTimeout) * time.Second,
		},
	}
	return nil
}

func (c *HttpClient) Close() {
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

// Do 核心请求方法
func (c *HttpClient) Do(ctx context.Context, method, url string, reqBody interface{}, opts ...RequestOption) ([]byte, error) {
	// 1. 序列化请求体
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// 2. 构建 Request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// 3. 应用 Option
	for _, opt := range opts {
		opt(req)
	}

	// 4. 执行请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// 5. 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// 6. HTTP 状态码转 error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return bodyBytes, httpStatusToError(resp.StatusCode, bodyBytes)
	}

	return bodyBytes, nil
}

// httpStatusToError 将 HTTP 状态码转为 error
func httpStatusToError(statusCode int, body []byte) error {
	bodyStr := string(body)
	switch {
	case statusCode == 400:
		return fmt.Errorf("bad request (400): %s", bodyStr)
	case statusCode == 401:
		return fmt.Errorf("unauthorized (401): %s", bodyStr)
	case statusCode == 403:
		return fmt.Errorf("forbidden (403): %s", bodyStr)
	case statusCode == 404:
		return fmt.Errorf("not found (404): %s", bodyStr)
	case statusCode >= 500:
		return fmt.Errorf("server error (%d): %s", statusCode, bodyStr)
	default:
		return fmt.Errorf("unexpected status code %d: %s", statusCode, bodyStr)
	}
}

// UnmarshalResponse 将 Do 返回的 bodyBytes 反序列化为指定类型
func unmarshalResponse[T any](bodyBytes []byte) (*T, error) {
	if len(bodyBytes) == 0 {
		return nil, nil
	}
	var data T
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &data, nil
}
