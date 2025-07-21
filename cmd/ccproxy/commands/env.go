package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// EnvCmd returns the env command
func EnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Show CCProxy environment variables",
		Long:  "Display the environment variables used by CCProxy and their current values",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🌍 CCProxy Environment Variables")
			fmt.Println()
			fmt.Println("Configuration:")
			fmt.Println("  CCPROXY_CONFIG      - Path to configuration file")
			fmt.Println("  CCPROXY_PORT        - Override default port (3456)")
			fmt.Println("  CCPROXY_HOST        - Override default host (127.0.0.1)")
			fmt.Println("  CCPROXY_API_KEY     - API key for authentication")
			fmt.Println()
			fmt.Println("Logging:")
			fmt.Println("  LOG                 - Enable file logging (true/false)")
			fmt.Println("  LOG_LEVEL           - Log level (debug, info, warn, error)")
			fmt.Println()
			fmt.Println("Claude Code Integration:")
			fmt.Println("  ANTHROPIC_BASE_URL  - Set by 'ccproxy code' command")
			fmt.Println("  ANTHROPIC_AUTH_TOKEN - Set by 'ccproxy code' command")
			fmt.Println("  API_TIMEOUT_MS      - Set by 'ccproxy code' command")
			fmt.Println()
			fmt.Println("Provider Configuration:")
			fmt.Println("  ANTHROPIC_API_KEY   - Anthropic API key")
			fmt.Println("  OPENAI_API_KEY      - OpenAI API key")
			fmt.Println("  GOOGLE_API_KEY      - Google AI API key")
			fmt.Println("  DEEPSEEK_API_KEY    - DeepSeek API key")
			fmt.Println("  OPENROUTER_API_KEY  - OpenRouter API key")
			fmt.Println("  GROQ_API_KEY        - Groq API key")
			fmt.Println()
			fmt.Println("Proxy Configuration:")
			fmt.Println("  HTTP_PROXY          - HTTP proxy URL")
			fmt.Println("  HTTPS_PROXY         - HTTPS proxy URL")
			fmt.Println("  NO_PROXY            - Hosts to bypass proxy")
		},
	}
}

