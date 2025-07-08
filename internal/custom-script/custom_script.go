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
	fmt.Printf("%s\n", jsonData)
	return nil
}

// outputTable outputs the result in table format
func outputTable(result *ScriptResult) error {
	fmt.Printf("┌─────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│                    CUSTOM SCRIPT EXECUTION RESULT              │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ Script Path: %-50s │\n", truncate(result.ScriptPath, 50))
	fmt.Printf("│ Status: %-56s │\n", result.Status)
	fmt.Printf("│ Exit Code: %-53d │\n", result.ExitCode)
	fmt.Printf("│ Execution Time: %-46.2f seconds │\n", result.ExecutionTime)
	fmt.Printf("│ Timestamp: %-51s │\n", result.TimestampUTC)
	fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")

	if result.ConfigsUsed.LimitsFile != "" {
		fmt.Printf("│ Limits File: %-49s │\n", truncate(result.ConfigsUsed.LimitsFile, 49))
	}
	if result.ConfigsUsed.RecommendationsFile != "" {
		fmt.Printf("│ Recommendations File: %-42s │\n", truncate(result.ConfigsUsed.RecommendationsFile, 42))
	}

	fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ OUTPUT                                                          │\n")
	fmt.Printf("├─────────────────────────────────────────────────────────────────┤\n")

	output := result.Output
	if result.Status == "FAIL" && result.ErrorOutput != "" {
		output = result.ErrorOutput
	}

	// Split output into lines and format for table
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if len(line) > 63 {
			// Split long lines
			for i := 0; i < len(line); i += 63 {
				end := i + 63
				if end > len(line) {
					end = len(line)
				}
				fmt.Printf("│ %-63s │\n", line[i:end])
			}
		} else {
			fmt.Printf("│ %-63s │\n", line)
		}
	}

	fmt.Printf("└─────────────────────────────────────────────────────────────────┘\n")
	return nil
}

// outputFriendly outputs the result in friendly format
func outputFriendly(result *ScriptResult) error {
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("🔧 CUSTOM SCRIPT EXECUTION RESULT\n")
	fmt.Printf(strings.Repeat("=", 70) + "\n")

	statusIcon := "✅"
	if result.Status == "FAIL" {
		statusIcon = "❌"
	}

	fmt.Printf("\n📊 EXECUTION SUMMARY:\n")
	fmt.Printf("   • Script: %s\n", result.ScriptPath)
	fmt.Printf("   • Status: %s %s\n", statusIcon, result.Status)
	fmt.Printf("   • Exit Code: %d\n", result.ExitCode)
	fmt.Printf("   • Execution Time: %.2f seconds\n", result.ExecutionTime)
	fmt.Printf("   • Timestamp: %s\n", result.TimestampUTC)

	if result.ConfigsUsed.LimitsFile != "" || result.ConfigsUsed.RecommendationsFile != "" {
		fmt.Printf("\n📁 CONFIGURATION FILES USED:\n")
		if result.ConfigsUsed.LimitsFile != "" {
			fmt.Printf("   • Limits File: %s\n", result.ConfigsUsed.LimitsFile)
		}
		if result.ConfigsUsed.RecommendationsFile != "" {
			fmt.Printf("   • Recommendations File: %s\n", result.ConfigsUsed.RecommendationsFile)
		}
	}

	fmt.Printf("\n" + strings.Repeat("-", 70) + "\n")
	fmt.Printf("📋 SCRIPT OUTPUT:\n")
	fmt.Printf(strings.Repeat("-", 70) + "\n")

	output := result.Output
	if result.Status == "FAIL" && result.ErrorOutput != "" {
		output = result.ErrorOutput
	}

	fmt.Printf("%s\n", output)

	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	return nil
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
