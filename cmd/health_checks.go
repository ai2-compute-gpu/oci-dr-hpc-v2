package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/recommender"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	skipRecommender bool
	tempResults     bool
)

var healthChecksCmd = &cobra.Command{
	Use:   "health-checks",
	Short: "Run comprehensive health checks with automated recommendations",
	Long: `Run all Level 1 diagnostic tests, generate a JSON report, and automatically analyze 
the results to provide intelligent recommendations for any issues found.

This command combines the functionality of 'level1' and 'recommender' commands for 
a streamlined health check experience.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting comprehensive health checks")

		// Initialize reporter
		rep := reporter.GetReporter()
		
		// Create temporary file for results if not specified
		var resultsFile string
		var cleanupFile bool
		
		outputFile := viper.GetString("output-file")
		if outputFile == "" {
			// Create temporary file
			tempFile, err := ioutil.TempFile("", "health-checks-*.json")
			if err != nil {
				return fmt.Errorf("failed to create temporary file: %w", err)
			}
			resultsFile = tempFile.Name()
			tempFile.Close()
			cleanupFile = tempResults // Only cleanup if --temp-results flag is set
		} else {
			resultsFile = outputFile
		}

		// Set append mode based on CLI flag
		appendMode := viper.GetBool("append")
		rep.SetAppendMode(appendMode)

		if err := rep.Initialize(resultsFile); err != nil {
			logger.Errorf("Failed to initialize reporter: %v", err)
			return fmt.Errorf("failed to initialize reporter: %w", err)
		}

		// Run all Level 1 tests
		if err := runLevel1HealthTests(); err != nil {
			logger.Errorf("Health checks failed: %v", err)
			// Continue to generate recommendations even if tests failed
		}

		// Generate JSON report
		logger.Info("Generating health check report...")
		if err := rep.WriteReportWithFormat("json"); err != nil {
			logger.Errorf("Failed to write health check report: %v", err)
			return fmt.Errorf("failed to write health check report: %w", err)
		}

		logger.Info(fmt.Sprintf("Health check report generated: %s", resultsFile))

		// Skip recommender if requested
		if skipRecommender {
			logger.Info("Skipping recommendation analysis (--skip-recommender flag)")
			if cleanupFile {
				os.Remove(resultsFile)
			}
			return nil
		}

		// Automatically run recommender analysis
		logger.Info("Analyzing results and generating recommendations...")
		
		// Get output format from configuration, default to friendly for better UX
		outputFormat := viper.GetString("output")
		if outputFormat == "" {
			outputFormat = "friendly"
		}

		if err := recommender.AnalyzeResults(resultsFile, outputFormat); err != nil {
			logger.Errorf("Failed to analyze results: %v", err)
			if cleanupFile {
				os.Remove(resultsFile)
			}
			return fmt.Errorf("failed to analyze results: %w", err)
		}

		// Cleanup temporary file if requested
		if cleanupFile {
			os.Remove(resultsFile)
			logger.Debug("Cleaned up temporary results file")
		}

		logger.Info("Health checks and recommendations completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthChecksCmd)
	healthChecksCmd.Flags().BoolVar(&skipRecommender, "skip-recommender", false, "skip automatic recommendation analysis")
	healthChecksCmd.Flags().BoolVar(&tempResults, "temp-results", true, "use temporary file for results (cleanup after analysis)")
}

// runLevel1HealthTests executes all Level 1 tests for health check
func runLevel1HealthTests() error {
	logger.Info("Running Level 1 health check tests")

	tests := GetLevel1Tests()

	var failedTests []string
	startTime := time.Now()

	for _, test := range tests {
		logger.Info(fmt.Sprintf("Running health check: %s", test.Name))
		if err := test.Fn(); err != nil {
			logger.Error(fmt.Sprintf("Health check %s failed: %v", test.Name, err))
			failedTests = append(failedTests, test.Name)
		}
	}

	duration := time.Since(startTime)
	logger.Info(fmt.Sprintf("Health checks completed in %v", duration))

	if len(failedTests) > 0 {
		logger.Warn(fmt.Sprintf("Health checks completed with %d issue(s): %v", len(failedTests), failedTests))
		return fmt.Errorf("health checks detected %d issue(s): %s", len(failedTests), strings.Join(failedTests, ", "))
	}

	logger.Info("All health checks passed successfully")
	return nil
}