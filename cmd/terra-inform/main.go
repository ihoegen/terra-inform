package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"math/rand"
	"time"
	"context"

	"github.com/ihoegen/terra-inform/pkg/checks"
	"github.com/ihoegen/terra-inform/pkg/provider"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	providerName = "openai"
	modelName    = "gpt-4o"
	aiProvider   provider.Provider
	// Define all available checks here
	allChecks = []checks.Check{
		checks.NewSummarizer(),
		// Add more checks here as we create them
	}
)

// parseArgs parses command line arguments, extracting our own flags and returning terraform commands
func parseArgs(args []string) ([]string, bool) {
	showHelp := false
	i := 1  // Skip program name
	tfArgs := []string{}

	for i < len(args) {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			showHelp = true
			i++
			continue
		} else if arg == "-p" || arg == "--provider" {
			if i+1 < len(args) {
				providerName = args[i+1]
				i += 2
				continue
			}
		} else if arg == "-m" || arg == "--model" {
			if i+1 < len(args) {
				modelName = args[i+1]
				i += 2
				continue
			}
		}

		// If we get here, it's a terraform argument
		tfArgs = append(tfArgs, arg)
		i++
	}

	return tfArgs, showHelp
}

func showHelpAndExit() {
	fmt.Println("terra-inform: An AI-powered wrapper for Terraform CLI")
	fmt.Println("\nUSAGE:")
	fmt.Println("  terra-inform [flags] <terraform commands>")
	fmt.Println("\nEXAMPLE:")
	fmt.Println("  terra-inform -m gpt-3.5-turbo plan")
	fmt.Println("  terra-inform apply")
	fmt.Println("\nFLAGS:")
	fmt.Println("  -h, --help                    Show this help message")
	fmt.Println("  -p, --provider <provider>     AI provider to use (default: openai)")
	fmt.Println("  -m, --model <model>           Model name to use (default: gpt-4o)")
	os.Exit(0)
}

// runAndAnalyzeCommand executes a terraform command and analyzes any errors
func runAndAnalyzeCommand(cmd *exec.Cmd, captureOutput bool) {
	var output strings.Builder
	if captureOutput {
		cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	} else {
		cmd.Stdout = os.Stdout
	}
	
	var errorOutput strings.Builder
	cmd.Stderr = io.MultiWriter(os.Stderr, &errorOutput)
	cmd.Stdin = os.Stdin
	
	err := cmd.Run()
	if err != nil {
		errorText := errorOutput.String()
		if errorText != "" {
			fmt.Println("\n🔍 Analyzing error...")
			// Use only the Summarizer check for error analysis
			errorCheck := checks.NewSummarizer()
			result, checkErr := analyzeWithCheck(errorCheck, errorText)
			if checkErr == nil {
				fmt.Printf("\n🤖 Error Analysis:\n%s\n", result)
			}
		}
		fmt.Printf("\nError running terraform: %v\n", err)
		os.Exit(1)
	}
	
	// If we need to capture output for AI analysis (like for plan/apply)
	if captureOutput && output.String() != "" {
		printAISummary(output.String())
	}
}

// analyzeWithCheck runs a single check and returns the result
func analyzeWithCheck(check checks.Check, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("no input to analyze")
	}
	
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)
	
	chatCompletion, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(check.GetPrompt(input)),
			}),
			Model:     openai.F(modelName),
			MaxTokens: openai.F(int64(500)),
		},
	)
	
	if err != nil {
		return "", fmt.Errorf("error analyzing with %s: %v", check.GetName(), err)
	}
	
	return chatCompletion.Choices[0].Message.Content, nil
}

func main() {
	// Parse arguments
	tfArgs, showHelp := parseArgs(os.Args)
	
	// Show help and exit if requested or no terraform commands
	if showHelp || len(tfArgs) == 0 {
		showHelpAndExit()
	}
	
	// Override with environment variables if set
	if envProvider := os.Getenv("terra-inform_MODEL_PROVIDER"); envProvider != "" {
		providerName = envProvider
	}
	if envModel := os.Getenv("terra-inform_MODEL_NAME"); envModel != "" {
		modelName = envModel
	}

	// Initialize provider
	switch providerName {
	case "openai":
		config := provider.Config{
			ModelName: modelName,
			APIKey:    os.Getenv("OPENAI_API_KEY"),
		}
		aiProvider = provider.NewOpenAIProvider(config)
	default:
		fmt.Printf("Unsupported provider: %s\n", providerName)
		os.Exit(1)
	}

	// Forward all arguments to terraform
	args := tfArgs

	// If it's a plan or apply command, we need to capture the output
	if args[0] == "plan" || args[0] == "apply" {
		// For apply without auto-approve, run plan first to get the changes
		if args[0] == "apply" && !contains(args, "-auto-approve") {
			// Generate random suffix for the plan file
			rand.Seed(time.Now().UnixNano())
			planFile := fmt.Sprintf("/tmp/tfplan-%d", rand.Int63())

			// Create plan file
			planCmd := exec.Command("terraform", "plan", "-out="+planFile)
			runAndAnalyzeCommand(planCmd, true)

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
			runAndAnalyzeCommand(cmd, false)

			// Clean up the plan file
			os.Remove(planFile)
		} else {
			// For plan or apply with auto-approve
			cmd := exec.Command("terraform", args...)
			runAndAnalyzeCommand(cmd, true)
		}
	} else {
		// For all other commands, just pass through to terraform
		cmd := exec.Command("terraform", args...)
		runAndAnalyzeCommand(cmd, false)
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

func printAISummary(planOutput string) {
	if planOutput == "" {
		return
	}
	fmt.Printf("\nRunning checks using %s Model: %s\n", providerName, modelName)
	
	results := aiProvider.ProcessChecks(allChecks, planOutput)

	fmt.Println("\n🤖 AI Analysis:")
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("\n❌ %s check failed: %v\n", result.CheckName, result.Error)
			continue
		}
		fmt.Printf("\n✅ %s:\n%s\n", result.CheckName, result.Result)
	}
} 