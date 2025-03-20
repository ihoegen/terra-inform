# terra-inform

terra-inform is a simple wrapper around the Terraform CLI that provides AI-powered summaries of your Terraform plans and applies. It works exactly like the regular Terraform CLI but adds helpful summaries to make infrastructure changes more understandable.

## Features

- Works as a drop-in replacement for the Terraform CLI
- Provides AI-generated summaries for `terraform plan` and `terraform apply` commands
- AI analysis of errors when Terraform commands fail
- Analyzes potential downtime risks from your infrastructure changes
- Maintains all standard Terraform functionality, including interactive approval for applies
- Forwards all other Terraform commands directly to the Terraform CLI
- Configurable OpenAI model selection via CLI flags or environment variables

## Prerequisites

- Terraform CLI installed and available in your PATH
- OpenAI API key

## Installation

1. Install Go if you haven't already:
   ```bash
   # macOS with Homebrew
   brew install go
   ```

2. Install the package:
   ```bash
   go install github.com/ihoegen/terra-inform/cmd/terra-inform@latest
   ```

## Configuration

### API Key

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY='your-api-key-here'
```

### Model Configuration

You can configure the AI provider and model either through environment variables or command-line flags.

#### Environment Variables

```bash
# Configure the AI model to use (defaults to "gpt-4o")
export TERRA_INFORM_MODEL_NAME='o3-mini'
```

#### Command-line Flags

```bash
# Specify the model
terra-inform -m o3-mini plan
```

Command-line flags take precedence over environment variables.

## Usage

Use terra-inform exactly as you would use the terraform command:

```bash
# Show help
terra-inform --help

# Instead of: terraform plan
terra-inform plan

# Instead of: terraform apply
terra-inform apply

# With model selection
terra-inform -m o3-mini plan

# Any other terraform commands work the same way
terra-inform init
terra-inform validate
terra-inform destroy
```

When running `plan` or `apply`, you'll get the standard Terraform output plus an AI-generated analysis that includes:

1. A comprehensive summary of the planned changes
2. A downtime risk assessment

When Terraform commands fail with errors, terra-inform will automatically analyze the error and provide insights to help you troubleshoot.

## License

MIT License 