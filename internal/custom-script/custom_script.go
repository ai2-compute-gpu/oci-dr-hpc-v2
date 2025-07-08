package custom_script

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/recommender"
	"github.com/oracle/oci-dr-hpc-v2/internal/test_limits"
	"github.com/spf13/viper"
)

// ScriptResult represents the result of executing a custom script
type ScriptResult struct {
	ScriptPath    string     `json:"script_path"`
	Status        string     `json:"status"`
	Output        string     `json:"output"`
	ErrorOutput   string     `json:"error_output,omitempty"`
	ExitCode      int        `json:"exit_code"`
	ExecutionTime float64    `json:"execution_time_seconds"`
	TimestampUTC  string     `json:"timestamp_utc"`
	ConfigsUsed   ConfigInfo `json:"configs_used"`
}

// ConfigInfo represents information about loaded configuration files
type ConfigInfo struct {
	LimitsFile          string `json:"limits_file,omitempty"`
	RecommendationsFile string `json:"recommendations_file,omitempty"`
}

// ExecuteScript executes a custom script with configuration support
func ExecuteScript(scriptPath, limitsFile, recommendationsFile, outputFormat string) error {
	logger.Infof("Executing custom script: %s", scriptPath)

	// Validate script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// Load configurations
	configInfo := ConfigInfo{}

	// Load test limits if specified
	if limitsFile != "" {
		if err := loadTestLimits(limitsFile); err != nil {
			logger.Infof("Warning: Failed to load test limits from %s: %v", limitsFile, err)
		} else {
			configInfo.LimitsFile = limitsFile
			logger.Infof("Loaded test limits from: %s", limitsFile)
		}
	}

	// Load recommendations if specified
	if recommendationsFile != "" {
		if err := loadRecommendations(recommendationsFile); err != nil {
			logger.Infof("Warning: Failed to load recommendations from %s: %v", recommendationsFile, err)
		} else {
			configInfo.RecommendationsFile = recommendationsFile
			logger.Infof("Loaded recommendations from: %s", recommendationsFile)
		}
	}

	// Execute the script
	result, err := executeScript(scriptPath, configInfo)
	if err != nil {
		return fmt.Errorf("failed to execute script: %w", err)
	}

	// Format and output the result
	if err := outputResult(result, outputFormat); err != nil {
		return fmt.Errorf("failed to output result: %w", err)
	}

	return nil
}

// executeScript runs the actual script and captures the results
func executeScript(scriptPath string, configInfo ConfigInfo) (*ScriptResult, error) {
	startTime := time.Now()

	// Determine script type and command
	ext := strings.ToLower(filepath.Ext(scriptPath))
	var cmd *exec.Cmd

	switch ext {
	case ".py":
		cmd = exec.Command("python3", scriptPath)
	case ".sh":
		cmd = exec.Command("bash", scriptPath)
	default:
		// Try to make it executable and run directly
		if err := os.Chmod(scriptPath, 0755); err != nil {
			logger.Infof("Warning: Failed to make script executable: %v", err)
		}
		cmd = exec.Command(scriptPath)
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Execute the command
	output, err := cmd.CombinedOutput()
	executionTime := time.Since(startTime).Seconds()

	result := &ScriptResult{
		ScriptPath:    scriptPath,
		ExecutionTime: executionTime,
		TimestampUTC:  time.Now().UTC().Format(time.RFC3339),
		ConfigsUsed:   configInfo,
	}

	if err != nil {
		result.Status = "FAIL"
		result.ErrorOutput = string(output)
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
		logger.Errorf("Script execution failed: %v", err)
	} else {
		result.Status = "PASS"
		result.Output = string(output)
		result.ExitCode = 0
		logger.Infof("Script executed successfully in %.2f seconds", executionTime)
	}

	return result, nil
}

// loadTestLimits loads test limits from the specified file
func loadTestLimits(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("test limits file not found: %s", filePath)
	}

	// Try to load the test limits
	_, err := test_limits.LoadTestLimitsFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load test limits: %w", err)
	}

	return nil
}

// loadRecommendations loads recommendations from the specified file
func loadRecommendations(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("recommendations file not found: %s", filePath)
	}

	// Try to parse the recommendations file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read recommendations file: %w", err)
	}

	var config recommender.RecommendationConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse recommendations file: %w", err)
	}

	return nil
}

// writeOutput writes content to either a file or stdout based on configuration
func writeOutput(content string) error {
	outputFile := viper.GetString("output-file")

	if outputFile != "" {
		// Write to file
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", outputFile, err)
		}
		logger.Infof("Output written to file: %s", outputFile)
	} else {
		// Write to stdout
		fmt.Print(content)
	}

	return nil
}

// outputResult formats and outputs the script result
func outputResult(result *ScriptResult, outputFormat string) error {
	switch outputFormat {
	case "json":
		return outputJSON(result)
	case "table":
		return outputTable(result)
	case "friendly":
		return outputFriendly(result)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// outputJSON outputs the result in JSON format
func outputJSON(result *ScriptResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return writeOutput(string(jsonData) + "\n")
}

// outputTable outputs the result in table format
func outputTable(result *ScriptResult) error {
	var output strings.Builder

	output.WriteString("┌─────────────────────────────────────────────────────────────────┐\n")
	output.WriteString("│                    CUSTOM SCRIPT EXECUTION RESULT              │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString(fmt.Sprintf("│ Script Path: %-50s │\n", truncate(result.ScriptPath, 50)))
	output.WriteString(fmt.Sprintf("│ Status: %-56s │\n", result.Status))
	output.WriteString(fmt.Sprintf("│ Exit Code: %-53d │\n", result.ExitCode))
	output.WriteString(fmt.Sprintf("│ Execution Time: %-46.2f seconds │\n", result.ExecutionTime))
	output.WriteString(fmt.Sprintf("│ Timestamp: %-51s │\n", result.TimestampUTC))
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	if result.ConfigsUsed.LimitsFile != "" {
		output.WriteString(fmt.Sprintf("│ Limits File: %-49s │\n", truncate(result.ConfigsUsed.LimitsFile, 49)))
	}
	if result.ConfigsUsed.RecommendationsFile != "" {
		output.WriteString(fmt.Sprintf("│ Recommendations File: %-42s │\n", truncate(result.ConfigsUsed.RecommendationsFile, 42)))
	}

	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")
	output.WriteString("│ OUTPUT                                                          │\n")
	output.WriteString("├─────────────────────────────────────────────────────────────────┤\n")

	scriptOutput := result.Output
	if result.Status == "FAIL" && result.ErrorOutput != "" {
		scriptOutput = result.ErrorOutput
	}

	// Split output into lines and format for table
	lines := strings.Split(strings.TrimSpace(scriptOutput), "\n")
	for _, line := range lines {
		if len(line) > 63 {
			// Split long lines
			for i := 0; i < len(line); i += 63 {
				end := i + 63
				if end > len(line) {
					end = len(line)
				}
				output.WriteString(fmt.Sprintf("│ %-63s │\n", line[i:end]))
			}
		} else {
			output.WriteString(fmt.Sprintf("│ %-63s │\n", line))
		}
	}

	output.WriteString("└─────────────────────────────────────────────────────────────────┘\n")
	return writeOutput(output.String())
}

// outputFriendly outputs the result in friendly format
func outputFriendly(result *ScriptResult) error {
	var output strings.Builder

	output.WriteString("\n" + strings.Repeat("=", 70) + "\n")
	output.WriteString("🔧 CUSTOM SCRIPT EXECUTION RESULT\n")
	output.WriteString(strings.Repeat("=", 70) + "\n")

	statusIcon := "✅"
	if result.Status == "FAIL" {
		statusIcon = "❌"
	}

	output.WriteString("\n📊 EXECUTION SUMMARY:\n")
	output.WriteString(fmt.Sprintf("   • Script: %s\n", result.ScriptPath))
	output.WriteString(fmt.Sprintf("   • Status: %s %s\n", statusIcon, result.Status))
	output.WriteString(fmt.Sprintf("   • Exit Code: %d\n", result.ExitCode))
	output.WriteString(fmt.Sprintf("   • Execution Time: %.2f seconds\n", result.ExecutionTime))
	output.WriteString(fmt.Sprintf("   • Timestamp: %s\n", result.TimestampUTC))

	if result.ConfigsUsed.LimitsFile != "" || result.ConfigsUsed.RecommendationsFile != "" {
		output.WriteString("\n📁 CONFIGURATION FILES USED:\n")
		if result.ConfigsUsed.LimitsFile != "" {
			output.WriteString(fmt.Sprintf("   • Limits File: %s\n", result.ConfigsUsed.LimitsFile))
		}
		if result.ConfigsUsed.RecommendationsFile != "" {
			output.WriteString(fmt.Sprintf("   • Recommendations File: %s\n", result.ConfigsUsed.RecommendationsFile))
		}
	}

	output.WriteString("\n" + strings.Repeat("-", 70) + "\n")
	output.WriteString("📋 SCRIPT OUTPUT:\n")
	output.WriteString(strings.Repeat("-", 70) + "\n")

	scriptOutput := result.Output
	if result.Status == "FAIL" && result.ErrorOutput != "" {
		scriptOutput = result.ErrorOutput
	}

	output.WriteString(fmt.Sprintf("%s\n", scriptOutput))
	output.WriteString("\n" + strings.Repeat("=", 70) + "\n")

	return writeOutput(output.String())
}

// truncate truncates a string to the specified length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// PrintMessage is kept for backward compatibility
func PrintMessage() {
	fmt.Println("Hello World - @rekharoy")
}
