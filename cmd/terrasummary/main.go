package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// getOpenAIModel returns the OpenAI model to use, either from TERRASUMMARY_MODEL env var or default to GPT-4
func getOpenAIModel() string {
	if model := os.Getenv("TERRASUMMARY_MODEL"); model != "" {
		return model
	}
	return "gpt-4o"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: terrasummary <terraform commands>")
		os.Exit(1)
	}

	// Forward all arguments to terraform
	args := os.Args[1:]

	// If it's a plan or apply command, we need to capture the output
	if args[0] == "plan" || args[0] == "apply" {
		var planOutput strings.Builder
		
		// For apply without auto-approve, run plan first to get the changes
		if args[0] == "apply" && !contains(args, "-auto-approve") {
			planCmd := exec.Command("terraform", "plan")
			planCmd.Stdout = &planOutput
			planCmd.Stderr = os.Stderr
			if err := planCmd.Run(); err != nil {
				fmt.Printf("Error running terraform plan: %v\n", err)
				os.Exit(1)
			}

			// Now run the actual apply command with full terminal interaction
			cmd := exec.Command("terraform", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin  // This is key - we pass through stdin directly
			
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running terraform apply: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For plan or apply with auto-approve
			cmd := exec.Command("terraform", args...)
			cmd.Stdout = io.MultiWriter(os.Stdout, &planOutput)
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running terraform: %v\n", err)
				os.Exit(1)
			}
		}

		// Generate summary using OpenAI
		if planOutput.Len() > 0 {
			fmt.Println("\nGenerating summary using OpenAI Model:", getOpenAIModel())
			summary := generateSummary(planOutput.String())
			fmt.Printf("\n🤖 AI Summary of Changes:\n%s\n", summary)
		}
	} else {
		// For all other commands, just pass through to terraform
		cmd := exec.Command("terraform", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running terraform: %v\n", err)
			os.Exit(1)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateSummary(planOutput string) string {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	chatCompletion, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a helpful assistant that summarizes Terraform plan output. " +
					"Focus on the key changes, resource additions, modifications, and deletions. " +
					"Be concise but comprehensive. Make sure this is easy to read and understand. " +
					"Format this output into a simple list with bullet points."),
				openai.UserMessage(planOutput),
			}),
			Model:     openai.F(getOpenAIModel()),
			MaxTokens: openai.F(int64(500)),
		},
	)

	if err != nil {
		return fmt.Sprintf("Error generating summary: %v", err)
	}

	return chatCompletion.Choices[0].Message.Content
} 