package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path-to-mcp-kokoro> [json-file]")
		os.Exit(1)
	}

	// Create the MCP client
	c, err := client.NewStdioMCPClient(os.Args[1], nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client
	fmt.Println("Initializing client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-kokoro-test",
		Version: "1.0.0",
	}

	initResult, err := c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	fmt.Printf("Initialized with server: %s %s\n\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// If a JSON file is provided, read and execute it
	if len(os.Args) > 2 {
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			log.Fatalf("Failed to read JSON file: %v", err)
		}

		var request mcp.CallToolRequest
		if err := json.Unmarshal(data, &request); err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}

		fmt.Printf("Calling tool: %s\n", request.Params.Name)
		result, err := c.CallTool(ctx, request)
		if err != nil {
			log.Fatalf("Tool call failed: %v", err)
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		// Interactive mode - list available tools
		toolsRequest := mcp.ListToolsRequest{}
		tools, err := c.ListTools(ctx, toolsRequest)
		if err != nil {
			log.Fatalf("Failed to list tools: %v", err)
		}

		fmt.Println("Available tools:")
		for _, tool := range tools.Tools {
			fmt.Printf("- %s: %s\n", tool.Name, tool.Description)
			if len(tool.InputSchema.Properties) > 0 {
				fmt.Printf("  Parameters:\n")
				for name := range tool.InputSchema.Properties {
					fmt.Printf("    - %s\n", name)
				}
			}
		}

		// List available prompts
		promptsRequest := mcp.ListPromptsRequest{}
		prompts, err := c.ListPrompts(ctx, promptsRequest)
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