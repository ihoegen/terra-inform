package provider

import (
	"sync"
	"github.com/ihoegen/terra-inform/pkg/checks"
)

// CheckResult holds the result of a check execution
type CheckResult struct {
	CheckName string
	Result    string
	Error     error
}

// ProcessFunction is the function type that providers must implement to process a single check
type ProcessFunction func(check checks.Check, input string) (string, error)

// Provider defines the interface that all AI providers must implement
type Provider interface {
	// ProcessChecks takes multiple checks and runs them in parallel, returning results in the same order as input checks
	ProcessChecks(checks []checks.Check, input string) []CheckResult
}

// Config holds the configuration for providers
type Config struct {
	ModelName string
	APIKey    string
}

// RunChecksInParallel executes checks in parallel using the provided process function
func RunChecksInParallel(checksToRun []checks.Check, input string, processFn ProcessFunction) []CheckResult {
	resultsChan := make(chan CheckResult, len(checksToRun))
	var wg sync.WaitGroup

	// Start all checks in parallel
	for _, check := range checksToRun {
		wg.Add(1)
		go func(c checks.Check) {
			defer wg.Done()
			result, err := processFn(c, input)
			resultsChan <- CheckResult{
				CheckName: c.GetName(),
				Result:    result,
				Error:     err,
			}
		}(check)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var results []CheckResult
	for result := range resultsChan {
		results = append(results, result)
	}

	sortedResults := make([]CheckResult, len(checksToRun))
	checkMap := make(map[string]CheckResult)
	for _, result := range results {
		checkMap[result.CheckName] = result
	}
	
	for i, check := range checksToRun {
		sortedResults[i] = checkMap[check.GetName()]
	}

	return sortedResults
} 