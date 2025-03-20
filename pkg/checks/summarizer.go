package checks

// Summarizer implements Check interface for summarizing terraform plan output
type Summarizer struct {
	BaseCheck
}

// NewSummarizer creates a new Summarizer check
func NewSummarizer() *Summarizer {
	return &Summarizer{
		BaseCheck: NewBaseCheck("summarizer"),
	}
}

// GetPrompt returns the prompt for the summarizer check
func (s *Summarizer) GetPrompt(planOutput string) string {
	return "You are a helpful assistant that summarizes Terraform plan output. " +
		"Focus on the key changes, resource additions, modifications, and deletions. " +
		"Be concise but comprehensive. Make sure this is easy to read and understand. " +
		"Format this output into a simple list with bullet points.\n\n" +
		planOutput
}