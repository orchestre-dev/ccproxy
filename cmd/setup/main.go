package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ProviderSetup represents configuration for a provider
type ProviderSetup struct {
	Name         string
	DisplayName  string
	Description  string
	DefaultModel string
	RequiredEnv  []EnvVar
	OptionalEnv  []EnvVar
}

// EnvVar represents an environment variable
type EnvVar struct {
	Key         string
	Description string
	Default     string
	Required    bool
}

var providers = []ProviderSetup{
	{
		Name:        "groq",
		DisplayName: "Groq",
		Description: "Ultra-fast inference with generous free tier",
		RequiredEnv: []EnvVar{
			{Key: "GROQ_API_KEY", Description: "Your Groq API key from console.groq.com", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "GROQ_MODEL", Description: "Model to use", Default: "moonshotai/kimi-k2-instruct"},
			{Key: "GROQ_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
		},
		DefaultModel: "moonshotai/kimi-k2-instruct",
	},
	{
		Name:        "openrouter",
		DisplayName: "OpenRouter",
		Description: "Access to 100+ models through a single API",
		RequiredEnv: []EnvVar{
			{Key: "OPENROUTER_API_KEY", Description: "Your OpenRouter API key from openrouter.ai", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "OPENROUTER_MODEL", Description: "Model to use", Default: "anthropic/claude-3.5-sonnet"},
			{Key: "OPENROUTER_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
			{Key: "OPENROUTER_SITE_URL", Description: "Your site URL for tracking", Default: ""},
			{Key: "OPENROUTER_SITE_NAME", Description: "Your site name for tracking", Default: ""},
		},
		DefaultModel: "anthropic/claude-3.5-sonnet",
	},
	{
		Name:        "openai",
		DisplayName: "OpenAI",
		Description: "Industry standard AI models with extensive tooling",
		RequiredEnv: []EnvVar{
			{Key: "OPENAI_API_KEY", Description: "Your OpenAI API key from platform.openai.com", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "OPENAI_MODEL", Description: "Model to use", Default: "gpt-4o"},
			{Key: "OPENAI_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
			{Key: "OPENAI_ORGANIZATION", Description: "Organization ID (optional)", Default: ""},
		},
		DefaultModel: "gpt-4o",
	},
	{
		Name:        "xai",
		DisplayName: "XAI (Grok)",
		Description: "Real-time information access and X/Twitter integration",
		RequiredEnv: []EnvVar{
			{Key: "XAI_API_KEY", Description: "Your XAI API key from console.x.ai", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "XAI_MODEL", Description: "Model to use", Default: "grok-beta"},
			{Key: "XAI_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
		},
		DefaultModel: "grok-beta",
	},
	{
		Name:        "gemini",
		DisplayName: "Google Gemini",
		Description: "Advanced multimodal AI with long context capabilities",
		RequiredEnv: []EnvVar{
			{Key: "GEMINI_API_KEY", Description: "Your Gemini API key from aistudio.google.com", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "GEMINI_MODEL", Description: "Model to use", Default: "gemini-1.5-flash"},
			{Key: "GEMINI_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
		},
		DefaultModel: "gemini-1.5-flash",
	},
	{
		Name:        "mistral",
		DisplayName: "Mistral AI",
		Description: "European AI with strong privacy focus and multilingual support",
		RequiredEnv: []EnvVar{
			{Key: "MISTRAL_API_KEY", Description: "Your Mistral API key from console.mistral.ai", Required: true},
		},
		OptionalEnv: []EnvVar{
			{Key: "MISTRAL_MODEL", Description: "Model to use", Default: "mistral-large-latest"},
			{Key: "MISTRAL_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
		},
		DefaultModel: "mistral-large-latest",
	},
	{
		Name:        "ollama",
		DisplayName: "Ollama",
		Description: "Local models with complete privacy and offline capabilities",
		RequiredEnv: []EnvVar{
			{Key: "OLLAMA_MODEL", Description: "Model to use (must be downloaded)", Required: true, Default: "llama3.2"},
		},
		OptionalEnv: []EnvVar{
			{Key: "OLLAMA_BASE_URL", Description: "Ollama server URL", Default: "http://localhost:11434"},
			{Key: "OLLAMA_MAX_TOKENS", Description: "Maximum tokens", Default: "16384"},
		},
		DefaultModel: "llama3.2",
	},
}

func main() {
	fmt.Println("🚀 CCProxy Setup Assistant")
	fmt.Println("==========================")
	fmt.Println()

	if len(os.Args) > 1 && os.Args[1] == "--help" {
		showHelp()
		return
	}

	// Check if .env already exists
	envPath := ".env"
	if _, err := os.Stat(envPath); err == nil {
		fmt.Printf("⚠️  Found existing %s file.\n", envPath)
		if !askConfirmation("Do you want to backup and replace it?") {
			fmt.Println("❌ Setup canceled.")
			return
		}

		// Backup existing .env
		backupPath := fmt.Sprintf(".env.backup.%d", os.Getpid())
		if err := os.Rename(envPath, backupPath); err != nil {
			fmt.Printf("❌ Failed to backup existing .env: %v\n", err)
			return
		}
		fmt.Printf("📦 Backed up existing .env to %s\n", backupPath)
		fmt.Println()
	}

	// Choose provider
	provider := chooseProvider()
	if provider == nil {
		fmt.Println("❌ Setup canceled.")
		return
	}

	fmt.Printf("Setting up %s provider...\n", provider.DisplayName)
	fmt.Println()

	// Collect configuration
	config := collectConfiguration(provider)

	// Generate .env file
	if err := generateEnvFile(provider, config); err != nil {
		fmt.Printf("❌ Failed to create .env file: %v\n", err)
		return
	}

	// Show completion message
	showCompletionMessage(provider)
}

func showHelp() {
	fmt.Println("CCProxy Setup Assistant")
	fmt.Println()
	fmt.Println("This tool helps you configure CCProxy for different AI providers.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ccproxy-setup          Run interactive setup")
	fmt.Println("  ccproxy-setup --help   Show this help message")
	fmt.Println()
	fmt.Println("Supported providers:")
	for _, p := range providers {
		fmt.Printf("  %-12s %s\n", p.Name, p.Description)
	}
	fmt.Println()
	fmt.Println("The setup will create a .env file with your configuration.")
	fmt.Println("Make sure to keep your API keys secure and never commit them to version control.")
}

func chooseProvider() *ProviderSetup {
	fmt.Println("📋 Available providers:")
	fmt.Println()
	for i, p := range providers {
		fmt.Printf("  %d. %s - %s\n", i+1, p.DisplayName, p.Description)
	}
	fmt.Println()

	for {
		fmt.Print("Choose a provider (1-7): ")
		input := readInput()

		if input == "" {
			continue
		}

		var choice int
		if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
			fmt.Println("❌ Please enter a number between 1 and 7")
			continue
		}

		if choice < 1 || choice > len(providers) {
			fmt.Println("❌ Please enter a number between 1 and 7")
			continue
		}

		return &providers[choice-1]
	}
}

func collectConfiguration(provider *ProviderSetup) map[string]string {
	config := make(map[string]string)

	fmt.Printf("🔧 Configuring %s\n", provider.DisplayName)
	fmt.Println()

	// Set provider
	config["PROVIDER"] = provider.Name

	// Collect required environment variables
	fmt.Println("Required configuration:")
	for _, env := range provider.RequiredEnv {
		for {
			fmt.Printf("  %s (%s): ", env.Key, env.Description)
			value := readInput()

			if value == "" && env.Default != "" {
				value = env.Default
			}

			if value == "" {
				fmt.Println("    ❌ This field is required")
				continue
			}

			config[env.Key] = value
			break
		}
	}

	fmt.Println()
	fmt.Println("Optional configuration (press Enter to use defaults):")

	// Collect optional environment variables
	for _, env := range provider.OptionalEnv {
		defaultText := ""
		if env.Default != "" {
			defaultText = fmt.Sprintf(" [default: %s]", env.Default)
		}

		fmt.Printf("  %s (%s)%s: ", env.Key, env.Description, defaultText)
		value := readInput()

		if value == "" && env.Default != "" {
			value = env.Default
		}

		if value != "" {
			config[env.Key] = value
		}
	}

	return config
}

func generateEnvFile(provider *ProviderSetup, config map[string]string) error {
	file, err := os.Create(".env")
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Write header
	if _, err := fmt.Fprintf(file, "# CCProxy Configuration - %s Provider\n", provider.DisplayName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "# Generated by CCProxy Setup Assistant\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "# %s\n\n", provider.Description); err != nil {
		return err
	}

	// Write main provider setting
	if _, err := fmt.Fprintf(file, "# Provider Selection\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "PROVIDER=%s\n\n", provider.Name); err != nil {
		return err
	}

	// Write provider-specific settings
	if _, err := fmt.Fprintf(file, "# %s Configuration\n", provider.DisplayName); err != nil {
		return err
	}

	// Write required vars first
	for _, env := range provider.RequiredEnv {
		if value, exists := config[env.Key]; exists {
			if _, err := fmt.Fprintf(file, "%s=%s\n", env.Key, value); err != nil {
				return err
			}
		}
	}

	// Write optional vars
	for _, env := range provider.OptionalEnv {
		if value, exists := config[env.Key]; exists {
			if _, err := fmt.Fprintf(file, "%s=%s\n", env.Key, value); err != nil {
				return err
			}
		}
	}

	// Add Claude Code integration settings
	if _, err := fmt.Fprintf(file, "\n# Claude Code Integration\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "# Uncomment these lines to use CCProxy with Claude Code\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "# export ANTHROPIC_BASE_URL=http://localhost:7187\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "# export ANTHROPIC_API_KEY=NOT_NEEDED\n"); err != nil {
		return err
	}

	return nil
}

func showCompletionMessage(provider *ProviderSetup) {
	fmt.Println()
	fmt.Println("✅ Setup completed successfully!")
	fmt.Println()
	fmt.Printf("📄 Created .env file with %s configuration.\n", provider.DisplayName)
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("1. Review your .env file and make any necessary adjustments")
	fmt.Println("2. Start CCProxy: ./ccproxy")
	fmt.Println("3. (Optional) Configure Claude Code to use CCProxy:")
	fmt.Println("   export ANTHROPIC_BASE_URL=http://localhost:7187")
	fmt.Println("   export ANTHROPIC_API_KEY=NOT_NEEDED")
	fmt.Println()

	if provider.Name == "ollama" {
		fmt.Println("📝 Ollama-specific notes:")
		fmt.Println("- Make sure Ollama is running: ollama serve")
		fmt.Printf("- Download the model: ollama pull %s\n", provider.DefaultModel)
		fmt.Println("- List available models: ollama list")
		fmt.Println()
	}

	fmt.Println("🔒 Security reminder:")
	fmt.Println("- Keep your API keys secure")
	fmt.Println("- Add .env to your .gitignore file")
	fmt.Println("- Never commit API keys to version control")
	fmt.Println()
	fmt.Println("📚 Documentation: https://your-docs-url.com")
	fmt.Println("🐛 Issues: https://github.com/your-repo/issues")
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(input)
}

func askConfirmation(question string) bool {
	fmt.Printf("%s (y/N): ", question)
	input := readInput()
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}

func init() {
	// Ensure we're in a directory where we can write files
	if _, err := os.Stat("."); err != nil {
		fmt.Printf("❌ Cannot access current directory: %v\n", err)
		os.Exit(1)
	}
}
