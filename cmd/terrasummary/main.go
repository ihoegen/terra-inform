package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"math/rand"
	"time"

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
			// Generate random suffix for the plan file
			rand.Seed(time.Now().UnixNano())
			planFile := fmt.Sprintf("/tmp/tfplan-%d", rand.Int63())

			// Create plan file
			planCmd := exec.Command("terraform", "plan", "-out="+planFile)
			planCmd.Stdout = io.MultiWriter(os.Stdout, &planOutput)
			planCmd.Stderr = os.Stderr
			if err := planCmd.Run(); err != nil {
				fmt.Printf("Error running terraform plan: %v\n", err)
				os.Exit(1)
			}

			// Generate and print summary before running apply
			printAISummary(planOutput.String())

			// Ask for confirmation
			fmt.Print("\nDo you want to perform these actions? Only 'yes' will be accepted to approve.\n\n")
			fmt.Print("Enter a value: ")

			var response string
			fmt.Scanln(&response)

			if response != "yes" {
				fmt.Println("Apply cancelled.")
				os.Remove(planFile)
				os.Exit(0)
			}

			// Now run apply with the saved plan file and auto-approve
			cmd := exec.Command("terraform", "apply", "-auto-approve", planFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running terraform apply: %v\n", err)
				os.Exit(1)
			}

			// Clean up the plan file
			os.Remove(planFile)
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

			printAISummary(planOutput.String())
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

func printAISummary(planOutput string) {
	if planOutput == "" {
		return
	}
	fmt.Println("\nGenerating summary using OpenAI Model:", getOpenAIModel())
	summary := generateSummary(planOutput)
	fmt.Printf("\n🤖 AI Summary of Changes:\n%s\n", summary)
} 