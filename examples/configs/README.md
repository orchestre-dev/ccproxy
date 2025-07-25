# CCProxy Configuration Examples

This directory contains ready-to-use configuration examples for CCProxy with popular OpenAI models.

## Available Configurations

### 1. `openai-gpt4.json` - GPT-4 Standard
Best for general-purpose use with the latest GPT-4 models.
- **Default**: GPT-4.1 (latest GPT-4 with improved coding and instruction following)
- **Background tasks**: GPT-4.1-mini (cost-effective for simple operations)
- **Long context**: GPT-4.1 (128K context window)

### 2. `openai-o-series.json` - O-Series Models
Optimized for reasoning and complex tasks.
- **Default**: GPT-4.1 (balanced performance)
- **Thinking tasks**: O3 (advanced reasoning)
- **Background tasks**: O4-mini (budget-friendly)
- Includes model mappings for Claude models to OpenAI equivalents

### 3. `openai-mixed.json` - Mixed Provider Setup
Combines OpenAI and Anthropic for optimal performance.
- **Default**: OpenAI GPT-4.1
- **Long context**: Anthropic Claude Opus 4 (best for very long documents)
- **Thinking**: OpenAI O3
- **Background**: OpenAI GPT-4.1-mini
- Includes performance tuning settings

### 4. `openai-budget.json` - Cost-Optimized
Maximum cost efficiency with mini models.
- All routes use cost-effective mini variants
- **Default**: GPT-4.1-mini
- **Thinking**: O1-mini
- **Background**: O4-mini

### 5. `qwen3-coder.json` - Qwen3-Coder Standalone
Alibaba's powerful 480B parameter coding model (July 2025).
- **Default**: qwen3-coder-plus (optimized for code generation)
- **Long context**: Up to 256K tokens native support
- **Background**: Same model with reduced token limits
- Excellent performance matching Claude-Sonnet-4

### 6. `qwen3-mixed.json` - Qwen3-Coder Multi-Provider
Combines Qwen3-Coder with other providers for flexibility.
- **Default**: Qwen3-Coder (cost-effective primary model)
- **Thinking**: Claude-Sonnet-4 for complex reasoning
- **Background**: GPT-4.1-turbo for quick tasks
- Maps Claude model requests to Qwen3-Coder for cost savings

### 7. `qwen3-budget.json` - Ultra Cost-Optimized with Qwen3-Coder
Maximum cost savings using Qwen3-Coder for all routes.
- All Claude model requests routed to Qwen3-Coder
- Optimized token limits for each use case
- Perfect for development and testing environments
- Significant cost reduction while maintaining quality

## Quick Start

1. Copy your desired configuration:
   ```bash
   cp examples/configs/openai-gpt4.json ~/.ccproxy/config.json
   ```

2. Add your OpenAI API key:
   ```bash
   # Set in the config file
   sed -i 's/YOUR_OPENAI_API_KEY/sk-your-actual-key/' ~/.ccproxy/config.json
   
   # Or use environment variable (recommended)
   export OPENAI_API_KEY="sk-your-actual-key"
   ```

3. Start CCProxy:
   ```bash
   ccproxy start
   ccproxy code  # Configure Claude Code environment
   ```

## Model Selection Guide

### GPT-4 Series
- **GPT-4.1**: Latest model with best overall performance
- **GPT-4.1-mini**: Cost-effective, 80% cheaper than GPT-4.1
- **GPT-4-turbo**: Previous generation, still powerful
- **GPT-4o**: Multimodal (text + vision)

### O-Series (Reasoning Models)
- **O3**: Best for complex reasoning and scientific tasks
- **O1**: Previous generation reasoning model
- **O1-mini**: Budget reasoning model
- **O4-mini**: Latest mini model for general tasks
- **O4-mini-high**: Enhanced mini model for code

### Qwen3-Coder Series
- **qwen3-coder-plus**: 480B parameter MoE model (35B active)
- **256K context window**: Native support, expandable to 1M
- **119 languages**: Comprehensive programming language support
- **69.6% SWE-bench**: Near parity with Claude-Sonnet-4 (70.4%)
- **Cost-effective**: Open-source with competitive pricing

## Routing Logic

CCProxy automatically routes requests based on:
1. **Token count > 60K**: Uses `longContext` route
2. **Haiku models**: Uses `background` route
3. **Thinking parameter**: Uses `think` route
4. **Explicit model mapping**: Direct model-to-model mapping
5. **Default**: Fallback for all other requests

## Tips

- Start with `openai-gpt4.json` for general use
- Use `openai-o-series.json` if you need advanced reasoning
- Choose `openai-budget.json` to minimize costs
- The `openai-mixed.json` provides the best of both worlds

## Environment Variables

You can override settings without modifying the config:
```bash
export CCPROXY_PORT=3456
export CCPROXY_HOST=127.0.0.1
export CCPROXY_API_KEY="your-auth-key"
export OPENAI_API_KEY="sk-openai-key"        # Auto-detected by provider name
export OPENROUTER_API_KEY="sk-or-v1-your-key" # For OpenRouter models including Qwen3-Coder
# Or use indexed format: export CCPROXY_PROVIDERS_0_API_KEY="sk-openai-key"
```