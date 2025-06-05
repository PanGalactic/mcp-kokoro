package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path-to-mcp-kokoro> [json-file]")
		os.Exit(1)
	}

	ctx := context.Background()
	mcpClient := client.NewStdioMCPClient(os.Args[1], nil)

	err := mcpClient.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer mcpClient.Close(ctx)

	// Initialize the connection
	_, err = mcpClient.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// If a JSON file is provided, read and execute it
	if len(os.Args) > 2 {
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			log.Fatalf("Failed to read JSON file: %v", err)
		}

		var request client.CallToolRequest
		if err := json.Unmarshal(data, &request); err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}

		result, err := mcpClient.CallTool(ctx, request)
		if err != nil {
			log.Fatalf("Tool call failed: %v", err)
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		// Interactive mode - list available tools
		tools, err := mcpClient.ListTools(ctx)
		if err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}

		fmt.Println("Available tools:")
		for _, tool := range tools.Tools {
			fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
		}

		// List available prompts
		prompts, err := mcpClient.ListPrompts(ctx)
		if err != nil {
			log.Fatalf("Failed to list prompts: %v", err)
		}

		if len(prompts.Prompts) > 0 {
			fmt.Println("\nAvailable prompts:")
			for _, prompt := range prompts.Prompts {
				fmt.Printf("- %s: %s\n", prompt.Name, prompt.Description)
			}
		}
	}
}