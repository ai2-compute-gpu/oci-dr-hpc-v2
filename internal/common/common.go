package common

// BaseTest Common test base struct
type BaseTest struct {
	name        string
	description string
}

// NewBaseTest creates a new BaseTest instance
func NewBaseTest(name, description string) *BaseTest {
	return &BaseTest{
		name:        name,
		description: description,
	}
}

// Name returns the name of the test
func (bt *BaseTest) Name() string {
	return bt.name
}

// Description returns the description of the test
func (bt *BaseTest) Description() string {
	return bt.description
}

// Run Runs the test and returns the result as raw text
func (bt *BaseTest) Run() (string, error) {
	// Placeholder for actual test logic
	// This should be overridden by specific test implementations
	return "Test executed successfully", nil
}
