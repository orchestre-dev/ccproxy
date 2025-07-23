#!/bin/bash

# This script demonstrates how CCProxy automatically detects provider API keys
# from environment variables

echo "🚀 CCProxy Environment Variable Demo"
echo "===================================="
echo ""
echo "This demo shows how CCProxy automatically detects provider-specific"
echo "environment variables without needing to specify them in config.json"
echo ""

# Set provider API keys using human-readable names
export ANTHROPIC_API_KEY="sk-ant-demo-key-123"
export OPENAI_API_KEY="sk-openai-demo-key-456"
export GEMINI_API_KEY="AI-gemini-demo-key-789"

echo "✅ Environment variables set:"
echo "   ANTHROPIC_API_KEY = ${ANTHROPIC_API_KEY}"
echo "   OPENAI_API_KEY = ${OPENAI_API_KEY}"
echo "   GEMINI_API_KEY = ${GEMINI_API_KEY}"
echo ""

echo "📄 Using simple config without API keys:"
cat simple-config.json
echo ""

echo "🔧 Starting CCProxy..."
echo "   The API keys will be automatically detected from environment variables"
echo ""
echo "   Run: ccproxy start --config simple-config.json"
echo ""

echo "💡 Benefits:"
echo "   - No API keys in config files"
echo "   - Human-readable environment variable names"
echo "   - Easy to manage in CI/CD and production"
echo "   - Backward compatible with indexed format"