# hm (Help Me) - AI-Powered CLI Assistant

`hm` is a command-line tool powered by Azure OpenAI that provides helpful explanations and suggestions for your CLI commands and tasks. Whether you're a beginner or an experienced user, `hm` can help you understand complex commands, remember forgotten syntax, and discover new solutions.

## Features

* **Explanations:**  Get clear and concise explanations of CLI commands.  Just ask `hm explain <command>`.
* **Suggestions:** Get helpful suggestions for CLI tasks.  Use `hm suggest <task>`.
* **Platform Awareness:** `hm` is aware of your operating system and likely shell, so it provides relevant and accurate information.
* **Customisable System Prompt:** Tailor the AI assistant's behaviour with a custom system prompt.
* **Configuration:**  Configure `hm` via command-line flags, environment variables, or a config file.

## Installation

### From Source (Go)

1. **Prerequisites:**  Ensure you have Go installed.  You'll also need the following libraries:

   ```bash
   go get github.com/spf13/cobra
   go get github.com/mhingston/azoai
   ```

2. **Clone the Repository:**

   ```bash
   git clone https://github.com/mhingston/hm
   ```

3. **Build:**

   ```bash
   go build hm.go
   ```

   Or use the build script: `build.sh`

## Configuration

`hm` supports configuration via command-line flags and a config file (`$HOME/.hm.json`).  Command-line flags take precedence, then the config file.

**Required Configuration:**

* **`api-key`:** Your Azure OpenAI API key.
* **`api-endpoint`:** Your Azure OpenAI API endpoint.
* **`deployment-id`:**  Your Azure OpenAI deployment ID.

**Optional Configuration:**

* **`system-prompt`:** A custom prompt to guide the model's behaviour. The default prompt includes platform and shell information.

**Config File Example (`$HOME/.hm.json`):**

```json
{
  "api-key": "YOUR_API_KEY",
  "api-endpoint": "YOUR_API_ENDPOINT",
  "deployment": "YOUR_DEPLOYMENT_ID",
  "system-prompt": "You are a helpful and concise command-line interface (CLI) assistant. You should provide clear and accurate explanations or suggestions for CLI commands and tasks. Prioritise commands and syntax appropriate for this platform and shell. If the user's query is unclear apologise that you aren't able to help." // Optional
}
```

## Usage

```bash
hm explain "ls -l"       # Explain the "ls -l" command
hm suggest "copy a file"  # Suggest a command to copy a file
```