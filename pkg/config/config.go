package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Config 总体配置结构（go-mcp-cli 格式）
type Config struct {
	Version string                   `json:"version"`
	Servers map[string]*ServerConfig `json:"servers"`
}

// ClaudeDesktopConfig Claude Desktop JSON 配置格式
type ClaudeDesktopConfig struct {
	MCPServers map[string]*ClaudeServerConfig `json:"mcpServers"`
}

// ClaudeServerConfig Claude Desktop 服务器配置
type ClaudeServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name       string            `json:"name"`
	Transport  string            `json:"transport"`
	Command    string            `json:"command,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	URL        string            `json:"url,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	MaxRetries int               `json:"maxRetries,omitempty"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	configPath string
	config     *Config
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath ...string) (*ConfigManager, error) {
	var path string
	if len(configPath) > 0 {
		path = configPath[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, ".mcp-cli", "config.json")
	}

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	cm := &ConfigManager{configPath: path}
	if err := cm.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if cm.config == nil {
		cm.config = &Config{
			Version: "1.0.0",
			Servers: make(map[string]*ServerConfig),
		}
	}

	return cm, nil
}

// load 加载配置文件
func (cm *ConfigManager) load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	// 尝试解析为 go-mcp-cli 格式
	var config Config
	if err := json.Unmarshal(data, &config); err == nil && config.Servers != nil {
		cm.config = &config
		return nil
	}

	// 尝试解析为 Claude Desktop 格式
	var claudeConfig ClaudeDesktopConfig
	if err := json.Unmarshal(data, &claudeConfig); err == nil && claudeConfig.MCPServers != nil {
		// 转换为 go-mcp-cli 格式
		cm.config = &Config{
			Version: "1.0.0",
			Servers: make(map[string]*ServerConfig),
		}
		for name, server := range claudeConfig.MCPServers {
			cm.config.Servers[name] = convertClaudeServer(name, server)
		}
		return nil
	}

	return fmt.Errorf("unsupported config format")
}

// LoadClaudeDesktopConfig 从 Claude Desktop 配置文件加载
func LoadClaudeDesktopConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var claudeConfig ClaudeDesktopConfig
	if err := json.Unmarshal(data, &claudeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Claude Desktop config: %w", err)
	}

	config := &Config{
		Version: "1.0.0",
		Servers: make(map[string]*ServerConfig),
	}
	for name, server := range claudeConfig.MCPServers {
		config.Servers[name] = convertClaudeServer(name, server)
	}

	return config, nil
}

// convertClaudeServer 将 Claude Desktop 配置转换为 go-mcp-cli 格式
func convertClaudeServer(name string, server *ClaudeServerConfig) *ServerConfig {
	config := &ServerConfig{
		Name:    name,
		Command: server.Command,
		Args:    server.Args,
		Env:     server.Env,
		Headers: server.Headers,
		URL:     server.URL,
	}

	// 自动检测传输类型
	if server.Command != "" {
		config.Transport = "stdio"
	} else if server.URL != "" {
		config.Transport = detectTransportType(server.URL)
	}

	return config
}

// detectTransportType 根据 URL 自动检测传输类型
func detectTransportType(url string) string {
	// 如果 URL 明确包含 /sse，使用 SSE
	if strings.Contains(url, "/sse") {
		return "sse"
	}
	// 默认使用 HTTP (streamable)
	return "http"
}

// save 保存配置文件
func (cm *ConfigManager) save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// AddServer 添加服务器配置
func (cm *ConfigManager) AddServer(name string, config *ServerConfig) error {
	if cm.config.Servers == nil {
		cm.config.Servers = make(map[string]*ServerConfig)
	}
	cm.config.Servers[name] = config
	return cm.save()
}

// RemoveServer 移除服务器配置
func (cm *ConfigManager) RemoveServer(name string) bool {
	if _, exists := cm.config.Servers[name]; !exists {
		return false
	}
	delete(cm.config.Servers, name)
	cm.save()
	return true
}

// GetServer 获取服务器配置
func (cm *ConfigManager) GetServer(name string) *ServerConfig {
	return cm.config.Servers[name]
}

// ListServers 列出所有服务器
func (cm *ConfigManager) ListServers() map[string]*ServerConfig {
	return cm.config.Servers
}

// GetServerNames 获取所有服务器名称
func (cm *ConfigManager) GetServerNames() []string {
	names := make([]string, 0, len(cm.config.Servers))
	for name := range cm.config.Servers {
		names = append(names, name)
	}
	return names
}

// ServerExists 检查服务器是否存在
func (cm *ConfigManager) ServerExists(name string) bool {
	_, exists := cm.config.Servers[name]
	return exists
}
