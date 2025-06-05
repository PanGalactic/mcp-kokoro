#!/bin/bash

# Test script for mcp-kokoro
# Run from the test directory

set -e

echo "Building mcp-kokoro..."
cd ..
go build -o mcp-kokoro
cd test

echo -e "\n=== Testing basic TTS ==="
go run main.go ../mcp-kokoro json/kokoro_basic.json

echo -e "\n=== Testing TTS with options ==="
go run main.go ../mcp-kokoro json/kokoro_with_options.json

echo -e "\n=== Listing available tools ==="
go run main.go ../mcp-kokoro

echo -e "\nAll tests completed!"