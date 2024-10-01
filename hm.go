package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mhingston/azoai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	APIKey       string `mapstructure:"api-key"`
	APIEndpoint  string `mapstructure:"api-endpoint"`
	APIVersion   string `mapstructure:"api-version"`
	Deployment   string `mapstructure:"deployment"`
	SystemPrompt string `mapstructure:"system-prompt"`
}

var config Config

func initConfig() {
	// Find home directory for config file
	homeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(homeDir)
	viper.SetConfigName(".hm")
	viper.SetConfigType("json")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Unmarshal config into struct (this will be overridden by flags later)
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config file: %s\n", err)
		os.Exit(1)
	}
}

func getCompletion(promptType, userInput string) (string, error) {
	// Get shell information (best effort)
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown shell"
	} else {
		shell = shell[strings.LastIndex(shell, "/")+1:] // Extract shell name (e.g., "bash", "zsh")
	}

	systemPrompt := config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful and concise command-line interface (CLI) assistant. You should provide clear and accurate explanations or suggestions for CLI commands and tasks. Prioritise commands and syntax appropriate for this platform and shell. If the user's query is unclear apologise that you aren't able to help."
	}

	systemPrompt = fmt.Sprintf("%s\n\nThe current platform is %s %s, likely using %s.", systemPrompt, runtime.GOOS, runtime.GOARCH, shell)

	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = "2024-06-01"
	}

	resp, err := azoai.InvokeOpenAIRequest(azoai.OpenAIRequest{
		SystemPrompt: systemPrompt,
		Message:      fmt.Sprintf("%s the following: %s", promptType, userInput),
		ApiBaseUrl:   config.APIEndpoint,
		APIKey:       config.APIKey,
		APIVersion:   apiVersion,
		Deployment:   config.Deployment,
	})

	return resp, err
}

func runExplain(cmd *cobra.Command, args []string) {
	userInput := strings.Join(args, " ")
	explanation, err := getCompletion("Explain the following command", userInput)
	cobra.CheckErr(err)
	fmt.Println(explanation)
}

func runSuggest(cmd *cobra.Command, args []string) {
	userInput := strings.Join(args, " ")
	suggestion, err := getCompletion("Suggest a command from the following description", userInput)
	cobra.CheckErr(err)
	fmt.Println(suggestion)
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd := &cobra.Command{
		Use:   "hm",
		Short: "Help Me - A CLI tool for explanations and suggestions",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Unmarshal config again after flags are processed
			if err := viper.Unmarshal(&config); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing config: %s\n", err)
				os.Exit(1)
			}

			// Check for required config values
			if config.APIKey == "" || config.APIEndpoint == "" || config.Deployment == "" {
				fmt.Fprintln(os.Stderr, "Missing required configuration values (API key, endpoint, or deployment ID). Please set them in the config file or using flags.")
				os.Exit(1)
			}
		},
	}

	explainCmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain a command or concept",
		Run:   runExplain,
	}

	suggestCmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest a command or solution",
		Run:   runSuggest,
	}

	rootCmd.PersistentFlags().StringVar(&config.APIKey, "api-key", "", "Azure OpenAI API key")
	rootCmd.PersistentFlags().StringVar(&config.APIEndpoint, "api-endpoint", "", "Azure OpenAI API endpoint")
	rootCmd.PersistentFlags().StringVar(&config.APIVersion, "api-version", "", "Azure OpenAI API version")
	rootCmd.PersistentFlags().StringVar(&config.Deployment, "deployment", "", "Azure OpenAI Deployment ID")
	rootCmd.PersistentFlags().StringVar(&config.SystemPrompt, "system-prompt", "", "System prompt")

	// Bind flags to Viper for config file support
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api-endpoint", rootCmd.PersistentFlags().Lookup("api-endpoint"))
	viper.BindPFlag("api-version", rootCmd.PersistentFlags().Lookup("api-version"))
	viper.BindPFlag("deployment", rootCmd.PersistentFlags().Lookup("deployment"))
	viper.BindPFlag("system-prompt", rootCmd.PersistentFlags().Lookup("system-prompt"))

	rootCmd.AddCommand(explainCmd, suggestCmd)

	cobra.CheckErr(rootCmd.Execute())
}
