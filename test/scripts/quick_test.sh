#!/bin/bash

# Quick command-line test for mcp-kokoro

echo "Testing mcp-kokoro with a simple TTS request..."

# Create a temporary JSON request
cat > /tmp/kokoro_test.json << 'EOF'
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "kokoro_tts",
    "arguments": {
      "text": "Testing Kokoro TTS from the command line"
    }
  }
}
EOF

# Send the request to mcp-kokoro via stdin and capture the response
echo "Sending request to mcp-kokoro..."
cat /tmp/kokoro_test.json | ~/go/bin/mcp-kokoro

# Clean up
rm /tmp/kokoro_test.json