package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Copyright 2025 MCP CLI Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// MCPClient 统一的 MCP 客户端包装
type MCPClient struct {
	client  *mcp.Client
	session *mcp.ClientSession
}

// StdioConfig Stdio 传输配置
type StdioConfig struct {
	Command string
	Args    []string
	Env     map[string]string
}

// SSEConfig SSE 传输配置
type SSEConfig struct {
	Endpoint   string            // SSE 端点 URL
	HTTPClient *http.Client      // HTTP 客户端（可选）
	Headers    map[string]string // 自定义请求头
}

// HTTPConfig HTTP 传输配置
type HTTPConfig struct {
	Endpoint   string            // HTTP 端点 URL
	HTTPClient *http.Client      // HTTP 客户端（可选）
	MaxRetries int               // 最大重试次数
	Logger     interface{}       // 日志记录器（可选）
	Headers    map[string]string // 自定义请求头
}

// NewClient 创建新的 MCP 客户端
func NewClient(name, version string) *MCPClient {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	return &MCPClient{
		client: client,
	}
}

// ConnectStdio 使用 stdio 传输连接到服务器
func (c *MCPClient) ConnectStdio(ctx context.Context, config *StdioConfig) error {
	cmd := exec.Command(config.Command, config.Args...)
	if len(config.Env) > 0 {
		env := slices.Clone(os.Environ()) // Clone to avoid modifying global os.Environ()
		for k, v := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	} else {
		cmd.Env = slices.Clone(os.Environ())
	}

	transport := &mcp.CommandTransport{Command: cmd}
	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.session = session
	return nil
}

// ConnectSSE 使用 SSE 传输连接到服务器
func (c *MCPClient) ConnectSSE(ctx context.Context, config *SSEConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("Endpoint is required for SSE transport")
	}
	transport := &mcp.SSEClientTransport{
		Endpoint: config.Endpoint,
	}
	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else if len(config.Headers) > 0 {
		httpClient = createHTTPClientWithHeaders(config.Headers)
	} else {
		httpClient = &http.Client{}
	}
	transport.HTTPClient = httpClient
	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.session = session
	return nil
}

// ConnectHTTP 使用 HTTP 传输连接到服务器
func (c *MCPClient) ConnectHTTP(ctx context.Context, config *HTTPConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("Endpoint is required for HTTP transport")
	}
	transport := &mcp.StreamableClientTransport{
		Endpoint:   config.Endpoint,
		MaxRetries: config.MaxRetries,
	}
	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else if len(config.Headers) > 0 {
		httpClient = createHTTPClientWithHeaders(config.Headers)
	} else {
		httpClient = &http.Client{}
	}
	transport.HTTPClient = httpClient
	session, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.session = session
	return nil
}

// createHTTPClientWithHeaders 创建一个带有自定义请求头的 HTTP 客户端
func createHTTPClientWithHeaders(headers map[string]string) *http.Client {
	return &http.Client{
		Transport: &headerTransport{headers: headers},
	}
}

// headerTransport 自定义 HTTP 传输，用于添加请求头
type headerTransport struct {
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// ListTools 列出所有可用工具
func (c *MCPClient) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}
	return c.session.ListTools(ctx, nil)
}

// CallTool 调用工具
func (c *MCPClient) CallTool(ctx context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}

	params := &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	}

	return c.session.CallTool(ctx, params)
}

// Close 关闭连接
func (c *MCPClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}

// IsConnected 检查是否已连接
func (c *MCPClient) IsConnected() bool {
	return c.session != nil
}
