package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type Config struct {
	APIKey       string `json:"api-key"`
	APIEndpoint  string `json:"api-endpoint"`
	DeploymentID string `json:"deployment-id"`
	SystemPrompt string `json:"system-prompt"`
}

var cfgFile string
var config Config

func loadConfig() {
	// Load config from file if specified
	if cfgFile != "" {
		configFile, err := os.ReadFile(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
			os.Exit(1)
		}

		err = json.Unmarshal(configFile, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing config file: %s\n", err)
			os.Exit(1)
		}
	}

	// Check for required config values
	if config.APIKey == "" || config.APIEndpoint == "" || config.DeploymentID == "" {
		fmt.Fprintln(os.Stderr, "Missing required configuration values (API key, endpoint, or deployment ID).")
		os.Exit(1)
	}
}

func callOpenAI(promptType, userInput string) (string, error) {
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-06-01", config.APIEndpoint, config.DeploymentID)

	// Get shell information (best effort)
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown shell"
	} else {
		shell = shell[strings.LastIndex(shell, "/")+1:] // Extract shell name (e.g., "bash", "zsh")
	}

	systemPrompt := config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful and concise command-line interface (CLI) assistant. You should provide clear and accurate explanations or suggestions for CLI commands and tasks. Prioritize commands and syntax appropriate for this platform and shell. If the user's query is unclear apologise that you aren't able to help."
	}

	systemPrompt = fmt.Sprintf("%s\n\nThe current platform is %s %s, likely using %s.", systemPrompt, runtime.GOOS, runtime.GOARCH, shell)

	requestBody := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf("%s the following: %s", promptType, userInput)},
		},
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestBodyJSON)))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content found in response")
}

func runExplain(cmd *cobra.Command, args []string) {
	userInput := strings.Join(args, " ")
	explanation, err := callOpenAI("Explain the following command", userInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(explanation)
}

func runSuggest(cmd *cobra.Command, args []string) {
	userInput := strings.Join(args, " ")
	suggestion, err := callOpenAI("Suggest a command from the following description", userInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)

	}
	fmt.Println(suggestion)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "hm",
		Short: "Help Me - A CLI tool for explanations and suggestions",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hm.json)")
	rootCmd.PersistentFlags().StringVar(&config.APIKey, "api-key", "", "Azure OpenAI API key")
	rootCmd.PersistentFlags().StringVar(&config.APIEndpoint, "api-endpoint", "", "Azure OpenAI API endpoint")
	rootCmd.PersistentFlags().StringVar(&config.DeploymentID, "deployment-id", "", "Azure OpenAI Deployment ID")
	rootCmd.PersistentFlags().StringVar(&config.SystemPrompt, "system-prompt", "", "System prompt")

	rootCmd.AddCommand(explainCmd, suggestCmd)

	// Find home directory for config file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		defaultCfgFile := filepath.Join(homeDir, ".hm.json")
		if cfgFile == "" { // Use default config file if none specified
			cfgFile = defaultCfgFile
		}
	}

	loadConfig() // Load config *after* flags are parsed
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

}
