package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/justinwongcn/go-mcp-cli/pkg/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-config <config-file>")
		os.Exit(1)
	}

	configPath := os.Args[1]

	claudeConfig, err := config.LoadClaudeDesktopConfig(configPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Successfully loaded Claude Desktop config")
	fmt.Printf("Version: %s\n", claudeConfig.Version)
	fmt.Printf("Total servers: %d\n\n", len(claudeConfig.Servers))

	for name, server := range claudeConfig.Servers {
		fmt.Printf("📦 Server: %s\n", name)
		fmt.Printf("   Transport: %s\n", server.Transport)
		if server.Command != "" {
			fmt.Printf("   Command: %s\n", server.Command)
			fmt.Printf("   Args: %v\n", server.Args)
		}
		if server.URL != "" {
			fmt.Printf("   URL: %s\n", server.URL)
		}
		if len(server.Headers) > 0 {
			headersJSON, _ := json.Marshal(server.Headers)
			fmt.Printf("   Headers: %s\n", string(headersJSON))
		}
		fmt.Println()
	}
}
