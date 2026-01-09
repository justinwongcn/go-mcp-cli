package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// Config 总体配置结构
type Config struct {
	Version string                   `json:"version"`
	Servers map[string]*ServerConfig `json:"servers"`
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
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, ".mcp-cli", "config.json")
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

	return json.Unmarshal(data, &cm.config)
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
