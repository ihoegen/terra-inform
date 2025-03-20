# terra-inform

terra-inform is a simple wrapper around the Terraform CLI that provides AI-powered summaries of your Terraform plans and applies. It works exactly like the regular Terraform CLI but adds helpful summaries to make infrastructure changes more understandable.

## Features

- Works as a drop-in replacement for the Terraform CLI
- Provides AI-generated summaries for `terraform plan` and `terraform apply` commands
- Maintains all standard Terraform functionality, including interactive approval for applies
- Forwards all other Terraform commands directly to the Terraform CLI
- Configurable OpenAI model selection

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

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY='your-api-key-here'
```

Optionally, configure the OpenAI model to use (defaults to GPT-4o):

```bash
export terra-inform_MODEL='o3-mini'  # Example: use GPT-3.5 Turbo instead
```

## Usage

Use terra-inform exactly as you would use the terraform command:

```bash
# Instead of: terraform plan
terra-inform plan

# Instead of: terraform apply
terra-inform apply

# Any other terraform commands work the same way
terra-inform init
terra-inform validate
terra-inform destroy
```

When running `plan` or `apply`, you'll get the standard Terraform output plus an AI-generated summary of the changes.

## License

MIT License 