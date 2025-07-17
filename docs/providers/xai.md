# XAI (Grok) Provider

XAI's Grok models offer real-time information access and cutting-edge AI capabilities, with unique access to X (formerly Twitter) data and current events.

## Why Choose XAI (Grok)?

- 📰 **Real-time data**: Access to current information and X/Twitter
- 🆕 **Cutting-edge**: Latest AI technology from Elon Musk's team
- 🔍 **Web search**: Built-in internet access capabilities
- 🚀 **Innovation**: Rapid development and new features

## Setup

### 1. Get an API Key

1. Visit [console.x.ai](https://console.x.ai)
2. Sign up for an account
3. Navigate to the API Keys section
4. Generate a new API key

### 2. Configure CCProxy

Set the following environment variables:

```bash
export PROVIDER=xai
export XAI_API_KEY=xai-your_xai_api_key_here
```

### 3. Optional Configuration

```bash
# Custom model (default: grok-beta)
export XAI_MODEL=grok-2-latest

# Custom max tokens (default: 16384)
export XAI_MAX_TOKENS=8192

# Custom base URL (default: https://api.x.ai/v1)
export XAI_BASE_URL=https://api.x.ai/v1
```

## Available Models

| Model | Context | Speed | Best For | Cost/1M tokens |
|-------|---------|-------|----------|----------------|
| **grok-beta** | 128K | ⚡⚡ | General use (default) | $5.00/$15.00 |
| **grok-2-latest** | 128K | ⚡⚡ | Latest features | $5.00/$15.00 |
| **grok-2-public** | 128K | ⚡⚡ | Stable version | $5.00/$15.00 |

### Model Capabilities

All Grok models include:
- **Real-time information**: Current events and news
- **X/Twitter integration**: Access to X platform data
- **Web search**: Built-in internet browsing
- **Vision capabilities**: Image understanding
- **Function calling**: Tool use support

## Pricing

### Current Pricing
- **Input tokens**: $5.00 per 1M tokens
- **Output tokens**: $15.00 per 1M tokens
- **No free tier** currently available

### Beta Pricing Notes
- Pricing may change as models move out of beta
- Current rates are competitive for the real-time capabilities
- Monitor [console.x.ai](https://console.x.ai) for updates

## Configuration Examples

### Basic Setup

```bash
# .env file
PROVIDER=xai
XAI_API_KEY=xai-your_api_key_here
```

### Latest Model Setup

```bash
# For newest features
PROVIDER=xai
XAI_API_KEY=xai-your_api_key_here
XAI_MODEL=grok-2-latest
XAI_MAX_TOKENS=16384
```

### Performance-Optimized Setup

```bash
# For faster responses
PROVIDER=xai
XAI_API_KEY=xai-your_api_key_here
XAI_MODEL=grok-beta
XAI_MAX_TOKENS=4096
```

## Usage with Claude Code

Once configured, use Claude Code normally:

```bash
# Set CCProxy as the API endpoint
export ANTHROPIC_BASE_URL=http://localhost:7187
export ANTHROPIC_API_KEY=NOT_NEEDED

# Use Claude Code with real-time capabilities
claude "What's the latest news about AI development?"
```

## Unique Features

### Real-Time Information

Grok models have built-in access to current information:

```bash
# Ask about current events
claude "What happened in tech news today?"

# Get stock market updates
claude "What's the current price of Tesla stock?"

# Check weather conditions
claude "What's the weather like in New York right now?"
```

### X/Twitter Integration

Access to X platform data:

```bash
# Analyze trending topics
claude "What's trending on X today?"

# Get social media insights
claude "What are people saying about the latest iPhone release?"
```

### Web Search Capabilities

Built-in internet browsing:

```bash
# Research topics
claude "Find the latest research papers on quantum computing"

# Compare products
claude "Compare the specifications of the latest laptops"
```

## Features

### ✅ Fully Supported
- Text generation
- Function calling
- Tool use
- Streaming responses
- Vision/image input
- Real-time information access
- Web search capabilities
- X/Twitter data access
- JSON mode
- Custom temperature/top_p

### ❌ Not Supported
- File uploads (beyond images)
- Custom model fine-tuning
- Embeddings generation

## Performance Tips

### 1. Leverage Real-Time Features

```bash
# Use Grok's strength for current information
claude "What are the latest developments in..."

# Ask for real-time data
claude "Current status of..."
```

### 2. Optimize for Speed

```bash
# Reduce max tokens for faster responses
export XAI_MAX_TOKENS=2048

# Use specific prompts
claude "Give me a brief update on..."
```

### 3. Take Advantage of Vision

```bash
# Analyze images with context
claude "What's happening in this image based on current events?"
```

## Use Cases

### 1. News and Current Events

```bash
# Get breaking news
claude "What are the top 3 news stories today?"

# Analyze trends
claude "What trends are emerging in the tech industry this week?"
```

### 2. Market Analysis

```bash
# Stock analysis
claude "Analyze the recent performance of tech stocks"

# Economic insights
claude "What are the current economic indicators saying?"
```

### 3. Social Media Monitoring

```bash
# Sentiment analysis
claude "What's the public sentiment about the new product launch?"

# Trend identification
claude "What topics are gaining traction on social media?"
```

### 4. Research with Current Data

```bash
# Academic research
claude "Find recent studies about climate change published this month"

# Competitive analysis
claude "What are competitors doing in the AI space recently?"
```

## Troubleshooting

### Rate Limit Errors

```json
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_error"
  }
}
```

**Solution**: XAI has rate limits. Implement retry logic with exponential backoff.

### Authentication Errors

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error"
  }
}
```

**Solution**: Verify your XAI API key is correct and active.

### Model Availability

```json
{
  "error": {
    "message": "Model not available",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Check [console.x.ai](https://console.x.ai) for model availability. Models may be in beta or temporarily unavailable.

### Content Restrictions

```json
{
  "error": {
    "message": "Content violates usage policy",
    "type": "invalid_request_error"
  }
}
```

**Solution**: Modify your request to comply with XAI's usage policies.

### Real-Time Data Limitations

Some real-time queries may fail:

```json
{
  "error": {
    "message": "Unable to access real-time data",
    "type": "service_unavailable"
  }
}
```

**Solution**: Rephrase your query or try again later.

## Best Practices

### 1. Leverage Unique Capabilities

```bash
# Use for time-sensitive information
# Combine with other providers for different strengths
# Take advantage of X/Twitter insights
```

### 2. Handle Real-Time Data Responsibly

```bash
# Verify important information from multiple sources
# Be aware that real-time data may be incomplete
# Use for trending topics and current awareness
```

### 3. Cost Management

```bash
# Monitor usage closely (no free tier)
# Use for tasks that benefit from real-time data
# Consider other providers for basic text generation
```

### 4. Prompt Engineering for Grok

```bash
# Be specific about timeframes: "today", "this week"
# Ask for sources when getting factual information
# Use for analysis of current events and trends
```

## Integration Examples

### Current Events Analysis

```python
# Using with Anthropic SDK via CCProxy
import anthropic

client = anthropic.Anthropic(
    api_key="NOT_NEEDED",
    base_url="http://localhost:7187"
)

response = client.messages.create(
    model="claude-3-sonnet",  # Maps to Grok via CCProxy
    messages=[{
        "role": "user", 
        "content": "What are the top tech news stories today and their implications?"
    }],
    max_tokens=1000
)
```

### Social Media Monitoring

```bash
# Monitor brand mentions
claude "What are people saying about our company on social media today?"

# Track competitor activity
claude "What new features did our competitors announce this week?"
```

### Market Research

```bash
# Get latest market data
claude "What are the current trends in the smartphone market?"

# Analyze consumer sentiment
claude "How are consumers reacting to the new product launch?"
```

## Monitoring

### Usage Tracking

```bash
# Monitor XAI usage at console.x.ai
# Check CCProxy logs for request patterns
tail -f ccproxy.log | grep xai

# Status endpoint
curl http://localhost:7187/status
```

### Cost Management

```bash
# Track costs carefully (no free tier)
# Monitor token usage
# Set usage alerts in XAI console
```

## Comparison with Other Providers

### When to Use XAI (Grok)

- ✅ Need current, real-time information
- ✅ Social media monitoring and analysis
- ✅ Breaking news and trend analysis
- ✅ Market research with current data

### When to Use Other Providers

- ❌ General text generation (use Groq for speed)
- ❌ Cost-sensitive applications (use free tiers)
- ❌ Historical data analysis (use any provider)
- ❌ Code generation (use specialized models)

## Future Developments

XAI is rapidly evolving:

- New model releases
- Enhanced real-time capabilities
- Improved X/Twitter integration
- Potential free tier introduction

Stay updated at [x.ai](https://x.ai) and [console.x.ai](https://console.x.ai).

## Next Steps

- Monitor [XAI announcements](https://x.ai) for new features
- Explore [real-time use cases](/guide/real-time) specific to Grok
- Compare costs with [other providers](/providers/) for your use case
- Set up [usage monitoring](/guide/monitoring) for cost control