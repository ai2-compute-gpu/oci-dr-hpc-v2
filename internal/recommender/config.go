package recommender

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

// RecommendationTemplate represents a recommendation template from config
type RecommendationTemplate struct {
	Type       string   `json:"type"`
	FaultCode  string   `json:"fault_code,omitempty"`
	Issue      string   `json:"issue"`
	Suggestion string   `json:"suggestion"`
	Commands   []string `json:"commands,omitempty"`
	References []string `json:"references,omitempty"`
}

// TestRecommendations represents recommendations for a specific test
type TestRecommendations struct {
	Fail *RecommendationTemplate `json:"fail,omitempty"`
	Pass *RecommendationTemplate `json:"pass,omitempty"`
}

// RecommendationConfig represents the entire recommendation configuration
type RecommendationConfig struct {
	Recommendations  map[string]TestRecommendations `json:"recommendations"`
	SummaryTemplates map[string]string              `json:"summary_templates"`
}

// LoadRecommendationConfig loads recommendation configuration from JSON file
func LoadRecommendationConfig() (*RecommendationConfig, error) {
	logger.Debugf("Starting recommendation config search...")

	// Look for config file in multiple locations (order matters - local override > user > system > development)
	configPaths := []string{
		"./recommendations.json", // Current directory (highest priority override)
	}

	// Add user home directory for user-specific configs
	if home, err := os.UserHomeDir(); err == nil {
		userConfigPath := filepath.Join(home, ".config/oci-dr-hpc/recommendations.json")
		configPaths = append(configPaths, userConfigPath)
		logger.Debugf("Added user config path: %s", userConfigPath)
	} else {
		logger.Debugf("Could not determine user home directory: %v", err)
	}

	// Add system locations
	configPaths = append(configPaths, []string{
		"/etc/oci-dr-hpc/recommendations.json",       // System config location
		"/usr/share/oci-dr-hpc/recommendations.json", // System data location
		"/etc/oci-dr-hpc-recommendations.json",       // Legacy location
		"configs/recommendations.json",               // Development location
	}...)

	logger.Debugf("Searching for recommendation config in %d locations: %v", len(configPaths), configPaths)

	var configData []byte
	var configFile string

	for i, path := range configPaths {
		logger.Debugf("Checking path %d/%d: %s", i+1, len(configPaths), path)

		if absPath, err := filepath.Abs(path); err == nil {
			logger.Debugf("  Absolute path: %s", absPath)

			if _, err := os.Stat(absPath); err == nil {
				logger.Debugf("  File exists, attempting to read...")

				if data, err := os.ReadFile(absPath); err == nil {
					configData = data
					configFile = absPath
					logger.Infof("Loading recommendation config from: %s", configFile)
					break
				} else {
					logger.Debugf("  Failed to read file: %v", err)
				}
			} else {
				logger.Debugf("  File does not exist: %v", err)
			}
		} else {
			logger.Debugf("  Failed to get absolute path: %v", err)
		}
	}

	if configData == nil {
		logger.Errorf("Recommendation config file not found in any of the searched locations: %v", configPaths)
		return nil, fmt.Errorf("recommendation config file not found in any of: %v", configPaths)
	}

	var config RecommendationConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse recommendation config: %w", err)
	}

	return &config, nil
}

// GetRecommendation generates a recommendation based on test result and config
func (config *RecommendationConfig) GetRecommendation(testName, status string, testResult TestResult) *Recommendation {
	testConfig, exists := config.Recommendations[testName]
	if !exists {
		logger.Errorf("No recommendation config found for test: %s", testName)
		return nil
	}

	var template *RecommendationTemplate
	switch strings.ToUpper(status) {
	case "FAIL":
		template = testConfig.Fail
	case "WARN":
		template = testConfig.Fail
	case "PASS":
		template = testConfig.Pass
	default:
		logger.Errorf("Unknown test status: %s", status)
		return nil
	}

	if template == nil {
		logger.Debugf("No template defined for %s status of test %s", status, testName)
		return nil
	}

	// Create recommendation and apply variable substitutions
	rec := &Recommendation{
		Type:       template.Type,
		TestName:   testName,
		FaultCode:  template.FaultCode,
		Issue:      applyVariableSubstitution(template.Issue, testResult),
		Suggestion: applyVariableSubstitution(template.Suggestion, testResult),
		Commands:   applyCommandSubstitutions(template.Commands, testResult),
		References: template.References,
	}

	return rec
}

// GetSummary generates a summary based on recommendation counts and templates
func (config *RecommendationConfig) GetSummary(totalIssues, criticalCount, warningCount int) string {
	if totalIssues == 0 {
		if template, exists := config.SummaryTemplates["no_issues"]; exists {
			return template
		}
		return "All diagnostic tests passed! Your HPC environment appears healthy."
	}

	if template, exists := config.SummaryTemplates["has_issues"]; exists {
		summary := strings.ReplaceAll(template, "{total_issues}", fmt.Sprintf("%d", totalIssues))
		summary = strings.ReplaceAll(summary, "{critical_count}", fmt.Sprintf("%d", criticalCount))
		summary = strings.ReplaceAll(summary, "{warning_count}", fmt.Sprintf("%d", warningCount))
		return summary
	}

	return fmt.Sprintf("Found %d issue(s) requiring attention: %d critical, %d warning",
		totalIssues, criticalCount, warningCount)
}

// applyVariableSubstitution replaces template variables with actual values
func applyVariableSubstitution(template string, testResult TestResult) string {
	result := template

	// Replace common variables
	result = strings.ReplaceAll(result, "{gpu_count}", fmt.Sprintf("%d", testResult.GPUCount))
	result = strings.ReplaceAll(result, "{num_rdma_nics}", fmt.Sprintf("%d", testResult.NumRDMANics))
	result = strings.ReplaceAll(result, "{failed_interfaces}", fmt.Sprintf("%s", testResult.FailedInterfaces))
	result = strings.ReplaceAll(result, "{max_uncorrectable}", fmt.Sprintf("%d", testResult.MaxUncorrectable))
	result = strings.ReplaceAll(result, "{max_correctable}", fmt.Sprintf("%d", testResult.MaxCorrectable))
	result = strings.ReplaceAll(result, "{eth0_present}", fmt.Sprintf("%t", testResult.Eth0Present))

	// Replace GPU mode check specific variables
	if len(testResult.EnabledGPUIndexes) > 0 {
		result = strings.ReplaceAll(result, "{enabled_gpu_indexes}", strings.Join(testResult.EnabledGPUIndexes, ","))
	} else {
		result = strings.ReplaceAll(result, "{enabled_gpu_indexes}", "")
	}

	return result
}

// applyCommandSubstitutions applies variable substitutions to command templates
func applyCommandSubstitutions(commands []string, testResult TestResult) []string {
	var result []string

	for _, cmd := range commands {
		substitutedCmd := applyVariableSubstitution(cmd, testResult)
		result = append(result, substitutedCmd)
	}

	return result
}
