package checks

// Check defines the interface that all terraform plan checks must implement
type Check interface {
	// GetPrompt returns the prompt to be used for the AI provider
	GetPrompt(planOutput string) string
	// GetName returns the name of the check
	GetName() string
}

// BaseCheck provides common functionality for checks
type BaseCheck struct {
	name string
}

// NewBaseCheck creates a new BaseCheck with the given name
func NewBaseCheck(name string) BaseCheck {
	return BaseCheck{
		name: name,
	}
}

func (b *BaseCheck) GetName() string {
	return b.name
} 