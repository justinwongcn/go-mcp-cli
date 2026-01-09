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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/justinwongcn/go-mcp-cli/pkg/client"
	"github.com/justinwongcn/go-mcp-cli/pkg/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-cli",
	Short: "MCP CLI tool for managing Model Context Protocol servers",
	Long: `A powerful CLI tool for managing and interacting with MCP servers.
Supports stdio, SSE, and Streamable HTTP transports.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	addTransport string
	addCommand   string
	addURL       string
	addArgs      []string
	addHeaders   []string
	addEnv       []string
	addRetries   int
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new MCP server configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		serverConfig := &config.ServerConfig{
			Name:      name,
			Transport: addTransport,
		}

		switch addTransport {
		case "stdio":
			if addCommand == "" {
				return fmt.Errorf("stdio transport requires --command")
			}
			serverConfig.Command = addCommand
			serverConfig.Args = addArgs
			serverConfig.Env = parseEnvVars(addEnv)

		case "sse", "http":
			if addURL == "" {
				return fmt.Errorf("%s transport requires --url", addTransport)
			}
			serverConfig.URL = addURL
			serverConfig.Headers = parseHeaders(addHeaders)
			if addTransport == "http" {
				serverConfig.MaxRetries = addRetries
			}

		default:
			return fmt.Errorf("unknown transport type: %s", addTransport)
		}

		if err := cm.AddServer(name, serverConfig); err != nil {
			return fmt.Errorf("failed to add server: %w", err)
		}

		fmt.Printf("✓ Added server: %s (%s)\n", name, addTransport)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addTransport, "transport", "t", "stdio", "Transport type (stdio, sse, http)")
	addCmd.Flags().StringVar(&addCommand, "command", "", "Command for stdio transport")
	addCmd.Flags().StringVar(&addURL, "url", "", "URL for SSE/HTTP transport")
	addCmd.Flags().StringArrayVar(&addArgs, "args", nil, "Arguments for command")
	addCmd.Flags().StringArrayVar(&addHeaders, "header", nil, "Headers for HTTP requests")
	addCmd.Flags().StringArrayVar(&addEnv, "env", nil, "Environment variables")
	addCmd.Flags().IntVar(&addRetries, "retries", 3, "Max retries for HTTP transport")
	rootCmd.AddCommand(addCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured servers",
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := config.NewConfigManager()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		names := cm.GetServerNames()
		if len(names) == 0 {
			fmt.Println("No MCP servers configured.")
			fmt.Println("Usage:")
			fmt.Println("  mcp-cli add <name> <transport> [options]")
			fmt.Println("  mcp-cli add time stdio --command uvx --args mcp-server-time")
			fmt.Println("  mcp-cli add context7 http --url https://mcp.context7.com/mcp")
			return
		}

		fmt.Println("Configured MCP Servers:")
		for _, name := range names {
			serverConfig := cm.GetServer(name)
			fmt.Printf("📦 %s (%s)\n", name, serverConfig.Transport)
			if serverConfig.Transport == "stdio" {
				args := ""
				if len(serverConfig.Args) > 0 {
					for _, arg := range serverConfig.Args {
						args += " " + arg
					}
				}
				fmt.Printf("   Command: %s%s\n", serverConfig.Command, args)
			} else {
				fmt.Printf("   URL: %s\n", serverConfig.URL)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a server configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cm, err := config.NewConfigManager()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		removed := cm.RemoveServer(name)
		if !removed {
			fmt.Printf("❌ Server not found: %s\n", name)
		} else {
			fmt.Printf("✓ Removed server: %s\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

var toolsCmd = &cobra.Command{
	Use:   "tools <server>",
	Short: "List available tools for a server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		serverName := args[0]

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		serverConfig := cm.GetServer(serverName)
		if serverConfig == nil {
			fmt.Printf("❌ Server not found: %s\n", serverName)
			return nil
		}

		cli := client.NewClient("mcp-cli", "1.0.0")
		defer cli.Close()

		// Connect based on transport type
		switch serverConfig.Transport {
		case "stdio":
			stdioConfig := &client.StdioConfig{
				Command: serverConfig.Command,
				Args:    serverConfig.Args,
				Env:     serverConfig.Env,
			}
			if err := cli.ConnectStdio(ctx, stdioConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		case "sse":
			sseConfig := &client.SSEConfig{
				Endpoint: serverConfig.URL,
			}
			if err := cli.ConnectSSE(ctx, sseConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		case "http":
			httpConfig := &client.HTTPConfig{
				Endpoint:   serverConfig.URL,
				MaxRetries: serverConfig.MaxRetries,
			}
			if err := cli.ConnectHTTP(ctx, httpConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		default:
			return fmt.Errorf("unknown transport type: %s", serverConfig.Transport)
		}

		fmt.Printf("\n📋 Tools for %s:\n\n", serverName)

		tools, err := cli.ListTools(ctx)
		if err != nil {
			return fmt.Errorf("failed to list tools: %w", err)
		}

		if len(tools.Tools) == 0 {
			fmt.Println("No tools available.")
			return nil
		}

		for i, tool := range tools.Tools {
			fmt.Printf("%d. %s\n", i+1, tool.Name)
			if tool.Description != "" {
				desc := tool.Description
				if len(desc) > 100 {
					desc = desc[:100] + "..."
				}
				fmt.Printf("   └─ %s\n", desc)
			}
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
}

var (
	callToolName string
	callArgs     []string
)

var callCmd = &cobra.Command{
	Use:   "call <server> <tool>",
	Short: "Call a tool on a server",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		serverName := args[0]
		toolName := args[1]

		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		serverConfig := cm.GetServer(serverName)
		if serverConfig == nil {
			fmt.Printf("❌ Server not found: %s\n", serverName)
			return nil
		}

		cli := client.NewClient("mcp-cli", "1.0.0")
		defer cli.Close()

		// Connect based on transport type
		switch serverConfig.Transport {
		case "stdio":
			stdioConfig := &client.StdioConfig{
				Command: serverConfig.Command,
				Args:    serverConfig.Args,
				Env:     serverConfig.Env,
			}
			if err := cli.ConnectStdio(ctx, stdioConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		case "sse":
			sseConfig := &client.SSEConfig{
				Endpoint: serverConfig.URL,
			}
			if err := cli.ConnectSSE(ctx, sseConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		case "http":
			httpConfig := &client.HTTPConfig{
				Endpoint:   serverConfig.URL,
				MaxRetries: serverConfig.MaxRetries,
			}
			if err := cli.ConnectHTTP(ctx, httpConfig); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

		default:
			return fmt.Errorf("unknown transport type: %s", serverConfig.Transport)
		}

		fmt.Printf("\n🔧 Calling %s on %s...\n\n", toolName, serverName)

		// Parse arguments
		argsMap := make(map[string]any)
		if len(callArgs) > 0 {
			for _, arg := range callArgs {
				parts := parseArg(arg)
				if len(parts) == 2 {
					argsMap[parts[0]] = parts[1]
				}
			}
		}

		result, err := cli.CallTool(ctx, toolName, argsMap)
		if err != nil {
			return fmt.Errorf("failed to call tool: %w", err)
		}

		// Format the result
		if result.IsError {
			fmt.Println("❌ Tool execution failed")
		}

		for _, content := range result.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				fmt.Println(text.Text)
			}
		}
		fmt.Println()

		return nil
	},
}

func init() {
	callCmd.Flags().StringArrayVarP(&callArgs, "arg", "a", nil, "Tool arguments (key=value)")
	rootCmd.AddCommand(callCmd)
}

// Helper functions

func parseEnvVars(envVars []string) map[string]string {
	result := make(map[string]string)
	for _, envVar := range envVars {
		parts := parseArg(envVar)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func parseHeaders(headers []string) map[string]string {
	result := make(map[string]string)
	for _, header := range headers {
		parts := parseArg(header)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func parseArg(arg string) []string {
	for i := 0; i < len(arg); i++ {
		if arg[i] == '=' {
			return []string{arg[:i], arg[i+1:]}
		}
	}
	return []string{arg}
}

// exec 命令 - 临时调用 MCP 工具，无需预先配置
var (
	execCommand  string
	execURL      string
	execArgs     []string
	execToolArgs []string
	execRetries  int
	execList     bool
)

var execCmd = &cobra.Command{
	Use:   "exec <transport> [tool]",
	Short: "Execute an MCP tool without pre-configuration",
	Long: `Execute an MCP tool directly from the command line without needing 
to add a server configuration first. Useful for one-off tool calls.

Examples:
  # Stdio transport - list tools
  mcp-cli exec stdio --list --command uvx --args mcp-server-time

  # Stdio transport - call a tool
  mcp-cli exec stdio get_current_time --command uvx --args mcp-server-time --arg timezone=Asia/Shanghai

  # SSE transport - list tools  
  mcp-cli exec sse --list --url https://gitmcp.io/modelcontextprotocol/typescript-sdk

  # SSE transport - call a tool
  mcp-cli exec sse search_typescript_sdk_code --url https://gitmcp.io/modelcontextprotocol/typescript-sdk --arg query=Client

  # HTTP transport - list tools
  mcp-cli exec http --list --url https://mcp.context7.com/mcp

  # HTTP transport - call a tool
  mcp-cli exec http resolve-library-id --url https://mcp.context7.com/mcp --arg query=Express --arg libraryName=express`,
	Args: func(cmd *cobra.Command, args []string) error {
		if execList {
			if len(args) < 1 {
				return fmt.Errorf("transport type is required")
			}
			return nil
		}
		if len(args) != 2 {
			return fmt.Errorf("requires 2 arguments (transport and tool), or use --list to list tools")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		transportType := args[0]
		toolName := ""
		if len(args) > 1 {
			toolName = args[1]
		}

		cli := client.NewClient("mcp-cli-exec", "1.0.0")
		defer cli.Close()

		var err error

		switch transportType {
		case "stdio":
			if execCommand == "" {
				return fmt.Errorf("stdio transport requires --command flag")
			}
			stdioConfig := &client.StdioConfig{
				Command: execCommand,
				Args:    execArgs,
			}
			err = cli.ConnectStdio(ctx, stdioConfig)

		case "sse":
			if execURL == "" {
				return fmt.Errorf("sse transport requires --url flag")
			}
			sseConfig := &client.SSEConfig{
				Endpoint: execURL,
			}
			err = cli.ConnectSSE(ctx, sseConfig)

		case "http":
			if execURL == "" {
				return fmt.Errorf("http transport requires --url flag")
			}
			httpConfig := &client.HTTPConfig{
				Endpoint:   execURL,
				MaxRetries: execRetries,
			}
			err = cli.ConnectHTTP(ctx, httpConfig)

		default:
			return fmt.Errorf("unknown transport type: %s (valid: stdio, sse, http)", transportType)
		}

		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		if execList {
			fmt.Printf("\n📋 Tools for %s server:\n\n", transportType)

			tools, err := cli.ListTools(ctx)
			if err != nil {
				return fmt.Errorf("failed to list tools: %w", err)
			}

			if len(tools.Tools) == 0 {
				fmt.Println("No tools available.")
				return nil
			}

			for i, tool := range tools.Tools {
				fmt.Printf("%d. %s\n", i+1, tool.Name)
				if tool.Description != "" {
					desc := tool.Description
					if len(desc) > 100 {
						desc = desc[:100] + "..."
					}
					fmt.Printf("   └─ %s\n", desc)
				}
			}
			fmt.Println()

			return nil
		}

		fmt.Printf("🔧 Executing %s on %s server...\n\n", toolName, transportType)

		// Parse tool arguments
		argsMap := make(map[string]any)
		for _, arg := range execToolArgs {
			parts := parseArg(arg)
			if len(parts) == 2 {
				argsMap[parts[0]] = parts[1]
			}
		}

		// Call the tool
		result, err := cli.CallTool(ctx, toolName, argsMap)
		if err != nil {
			return fmt.Errorf("failed to call tool: %w", err)
		}

		// Format the result
		if result.IsError {
			fmt.Println("❌ Tool execution failed")
		}

		for _, content := range result.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				fmt.Println(text.Text)
			}
		}
		fmt.Println()

		return nil
	},
}

func init() {
	execCmd.Flags().StringVar(&execCommand, "command", "", "Command for stdio transport")
	execCmd.Flags().StringVar(&execURL, "url", "", "URL for SSE/HTTP transport")
	execCmd.Flags().StringArrayVar(&execArgs, "args", nil, "Arguments for command (stdio transport)")
	execCmd.Flags().StringArrayVarP(&execToolArgs, "arg", "a", nil, "Tool arguments (key=value)")
	execCmd.Flags().IntVar(&execRetries, "retries", 3, "Max retries for HTTP transport")
	execCmd.Flags().BoolVar(&execList, "list", false, "List available tools without calling a specific tool")
	rootCmd.AddCommand(execCmd)
}
