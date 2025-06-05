package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "mcp-kokoro",
	Short: "An MCP server for Kokoro TTS API",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetLevel(log.ErrorLevel)
}

type TTSRequest struct {
	Text  string  `json:"text"`
	Voice string  `json:"voice,omitempty"`
	Speed float64 `json:"speed,omitempty"`
}

func runServer() error {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Kokoro TTS Service",
		Version,
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	kokoroURL := os.Getenv("KOKORO_URL")
	if kokoroURL == "" {
		kokoroURL = "http://192.168.0.253:44444"
	}

	// Create the Kokoro TTS tool
	kokoroTool := mcp.NewTool("kokoro_tts",
		mcp.WithDescription("Uses the Kokoro TTS API to generate speech from text"),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The text message to convert to speech"),
		),
		mcp.WithString("voice",
			mcp.Description("Voice to use (optional)"),
		),
		mcp.WithNumber("speed",
			mcp.Description("Speed of speech (optional, default: 1.0)"),
		),
	)

	// Add the tool handler
	s.AddTool(kokoroTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		
		text, ok := arguments["text"].(string)
		if !ok || text == "" {
			result := mcp.NewToolResultText("Error: text is required and must be a string")
			result.IsError = true
			return result, nil
		}

		speed := 1.0
		if s, ok := arguments["speed"].(float64); ok {
			speed = s
		}

		voice := ""
		if v, ok := arguments["voice"].(string); ok {
			voice = v
		}

		ttsReq := TTSRequest{
			Text:  text,
			Voice: voice,
			Speed: speed,
		}

		jsonData, err := json.Marshal(ttsReq)
		if err != nil {
			result := mcp.NewToolResultText(fmt.Sprintf("Error: failed to marshal request: %v", err))
			result.IsError = true
			return result, nil
		}

		req, err := http.NewRequestWithContext(ctx, "POST", kokoroURL+"/voice/generate", bytes.NewBuffer(jsonData))
		if err != nil {
			result := mcp.NewToolResultText(fmt.Sprintf("Error: failed to create request: %v", err))
			result.IsError = true
			return result, nil
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			result := mcp.NewToolResultText(fmt.Sprintf("Error: failed to call Kokoro API: %v", err))
			result.IsError = true
			return result, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			result := mcp.NewToolResultText(fmt.Sprintf("Error: Kokoro API error (status %d): %s", resp.StatusCode, string(body)))
			result.IsError = true
			return result, nil
		}

		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			result := mcp.NewToolResultText(fmt.Sprintf("Error: failed to read audio data: %v", err))
			result.IsError = true
			return result, nil
		}

		encoded := base64.StdEncoding.EncodeToString(audioData)
		
		return mcp.NewToolResultText(fmt.Sprintf("Audio generated successfully (%d bytes). Base64 encoded audio data:\n%s", len(audioData), encoded)), nil
	})

	// Add prompts
	s.AddPrompt(mcp.NewPrompt("kokoro_tts_examples",
		mcp.WithPromptDescription("Example usage of Kokoro TTS with various options"),
		mcp.WithArgument("example_type",
			mcp.ArgumentDescription("Type of example: basic, voices, speed"),
		),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		exampleType := request.Params.Arguments["example_type"]
		
		examples := map[string]string{
			"basic": "Basic usage:\n```json\n{\n  \"text\": \"Hello, this is a test of Kokoro TTS.\"\n}\n```",
			"voices": "Using different voices:\n```json\n{\n  \"text\": \"Testing different voice options\",\n  \"voice\": \"female_1\"\n}\n```",
			"speed": "Adjusting speech speed:\n```json\n{\n  \"text\": \"This text will be spoken at different speeds\",\n  \"speed\": 1.5\n}\n```",
		}

		content := examples[exampleType]
		if content == "" {
			content = examples["basic"] + "\n\n" + examples["voices"] + "\n\n" + examples["speed"]
		}

		return mcp.NewGetPromptResult(
			"Examples of using Kokoro TTS",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(content),
				),
			},
		), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Error("Server error", "error", err)
		return err
	}

	return nil
}