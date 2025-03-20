package checks

// DowntimeAnalyzer implements Check interface for analyzing potential downtime
type DowntimeAnalyzer struct {
	BaseCheck
}

// NewDowntimeAnalyzer creates a new DowntimeAnalyzer check
func NewDowntimeAnalyzer() *DowntimeAnalyzer {
	return &DowntimeAnalyzer{
		BaseCheck: NewBaseCheck("downtime-analyzer"),
	}
}

// GetPrompt returns the prompt for the downtime analyzer check
func (d *DowntimeAnalyzer) GetPrompt(planOutput string) string {
	return "You are a Terraform expert focused on identifying potential downtime or service disruptions. " +
		"Analyze the following Terraform plan and identify if there are any changes that could cause " +
		"downtime or service disruption. Consider resource replacements, restarts, " +
		"or changes to critical infrastructure components like load balancers, databases, and networking. " +
		"Be extremely concise - respond with a single line starting with 'DOWNTIME RISK:' and a rating of None, Low, Medium, or High, " +
		"followed by a brief explanation if there is risk. Example: 'DOWNTIME RISK: Medium - Database instance will be restarted.'\n\n" +
		planOutput
} 