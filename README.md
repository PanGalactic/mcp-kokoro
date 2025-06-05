# mcp-kokoro

An MCP (Model Context Protocol) server for Kokoro TTS API integration.

## Features

- Text-to-speech generation using Kokoro API
- Support for different voices and speech speeds
- Base64 encoded audio output
- MCP protocol compliance

## Installation

```bash
go install github.com/yourusername/mcp-kokoro@latest
```

## Configuration

Set the Kokoro API URL (optional, defaults to http://192.168.0.253:44444):

```bash
export KOKORO_URL=http://your-kokoro-api-url:port
```

## Usage

### With Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "kokoro": {
      "command": "mcp-kokoro",
      "env": {
        "KOKORO_URL": "http://192.168.0.253:44444"
      }
    }
  }
}
```

### Available Tools

#### kokoro_tts
Generates speech from text using Kokoro TTS API.

Parameters:
- `text` (required): The text message to convert to speech
- `voice` (optional): Voice to use
- `speed` (optional): Speed of speech (default: 1.0)

Example:
```json
{
  "text": "Hello, this is a test of Kokoro TTS.",
  "voice": "female_1",
  "speed": 1.2
}
```

## Development

### Building

```bash
go build -o mcp-kokoro
```

### Testing

```bash
cd test
go run main.go ../mcp-kokoro
```

## License

MIT