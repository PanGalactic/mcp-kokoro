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
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-kokoro",
	Short: "An MCP server for Kokoro TTS API",
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
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

func runServer() {
	s := server.NewMCPServer(
		"mcp-kokoro",
		"0.1.0",
		server.WithPrompts(getPrompts()...),
	)

	kokoroURL := os.Getenv("KOKORO_URL")
	if kokoroURL == "" {
		kokoroURL = "http://192.168.0.253:44444"
	}

	s.AddTool(mcp.NewTool(
		"kokoro_tts",
		"Uses the Kokoro TTS API to generate speech from text",
		mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "The text message to convert to speech",
				},
				"voice": map[string]interface{}{
					"type":        "string",
					"description": "Voice to use (optional)",
				},
				"speed": map[string]interface{}{
					"type":        "number",
					"description": "Speed of speech (optional, default: 1.0)",
				},
			},
			Required: []string{"text"},
		},
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Text  string  `json:"text"`
				Voice string  `json:"voice"`
				Speed float64 `json:"speed"`
			}

			if err := json.Unmarshal(request.Params.Arguments, &args); err != nil {
				return mcp.NewToolResultError(fmt.Errorf("invalid arguments: %w", err)), nil
			}

			if args.Text == "" {
				return mcp.NewToolResultError(fmt.Errorf("text is required")), nil
			}

			if args.Speed == 0 {
				args.Speed = 1.0
			}

			ttsReq := TTSRequest{
				Text:  args.Text,
				Voice: args.Voice,
				Speed: args.Speed,
			}

			jsonData, err := json.Marshal(ttsReq)
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to marshal request: %w", err)), nil
			}

			req, err := http.NewRequestWithContext(ctx, "POST", kokoroURL+"/voice/generate", bytes.NewBuffer(jsonData))
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to create request: %w", err)), nil
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to call Kokoro API: %w", err)), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return mcp.NewToolResultError(fmt.Errorf("Kokoro API error (status %d): %s", resp.StatusCode, string(body))), nil
			}

			audioData, err := io.ReadAll(resp.Body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Errorf("failed to read audio data: %w", err)), nil
			}

			encoded := base64.StdEncoding.EncodeToString(audioData)
			
			return mcp.NewToolResultText(fmt.Sprintf("Audio generated successfully (%d bytes). Base64 encoded audio data:\n%s", len(audioData), encoded)), nil
		},
	))

	if err := s.Serve(server.StdioServerTransport()); err != nil {
		log.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func getPrompts() []mcp.Prompt {
	return []mcp.Prompt{
		{
			Name:        "kokoro_tts_examples",
			Description: "Example usage of Kokoro TTS with various options",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "example_type",
					Description: "Type of example: basic, voices, speed",
					Required:    false,
				},
			},
			GetPrompt: func(args map[string]string) (*mcp.GetPromptResult, error) {
				exampleType := args["example_type"]
				
				examples := map[string]string{
					"basic": "Basic usage:\n```json\n{\n  \"text\": \"Hello, this is a test of Kokoro TTS.\"\n}\n```",
					"voices": "Using different voices:\n```json\n{\n  \"text\": \"Testing different voice options\",\n  \"voice\": \"female_1\"\n}\n```",
					"speed": "Adjusting speech speed:\n```json\n{\n  \"text\": \"This text will be spoken at different speeds\",\n  \"speed\": 1.5\n}\n```",
				}

				content := examples[exampleType]
				if content == "" {
					content = strings.Join([]string{examples["basic"], examples["voices"], examples["speed"]}, "\n\n")
				}

				return &mcp.GetPromptResult{
					Description: "Examples of using Kokoro TTS",
					Messages: []mcp.SamplingMessage{
						{
							Role:    mcp.RoleUser,
							Content: mcp.NewTextContent(content),
						},
					},
				}, nil
			},
		},
	}
}