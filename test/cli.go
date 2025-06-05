package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	// Parse command line arguments
	text := flag.String("text", "Hello from Kokoro TTS", "Text to convert to speech")
	voice := flag.String("voice", "", "Voice to use (optional)")
	speed := flag.Float64("speed", 1.0, "Speech speed (optional)")
	playAudio := flag.Bool("play", true, "Play audio through speakers")
	flag.Parse()

	// Create the MCP client
	c, err := client.NewStdioMCPClient(os.Getenv("HOME")+"/go/bin/mcp-kokoro", nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp-kokoro-cli",
		Version: "1.0.0",
	}

	_, err = c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Prepare the tool call
	args := map[string]interface{}{
		"text": *text,
		"play_audio": *playAudio,
	}
	if *voice != "" {
		args["voice"] = *voice
	}
	if *speed != 1.0 {
		args["speed"] = *speed
	}

	// Call the kokoro_tts tool
	request := mcp.CallToolRequest{}
	request.Params.Name = "kokoro_tts"
	request.Params.Arguments = args

	fmt.Printf("Calling Kokoro TTS with text: %s\n", *text)
	result, err := c.CallTool(ctx, request)
	if err != nil {
		log.Fatalf("Tool call failed: %v", err)
	}

	// Print the result
	if result.IsError {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			fmt.Println("Error:", textContent.Text)
		}
	} else {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			fmt.Println("Success:", textContent.Text)
		}
	}
}