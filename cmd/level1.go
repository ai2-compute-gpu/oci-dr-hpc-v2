package cmd

import (
	"fmt"
	"strings"

	"github.com/oracle/oci-dr-hpc-v2/internal/level1_tests"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/oracle/oci-dr-hpc-v2/internal/reporter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	testFilter string
	listTests  bool
)

var level1Cmd = &cobra.Command{
	Use:          "level1",
	Short:        "Run Level 1 diagnostic tests",
	Long:         `Run Level 1 diagnostic tests for HPC environment. Use --test flag to run specific tests.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting Level 1 diagnostic tests")

		// Initialize reporter
		rep := reporter.GetReporter()
		outputFile := viper.GetString("output-file")

		// Set append mode based on CLI flag
		appendMode := viper.GetBool("append")
		rep.SetAppendMode(appendMode)

		if err := rep.Initialize(outputFile); err != nil {
			logger.Errorf("Failed to initialize reporter: %v", err)
			return fmt.Errorf("failed to initialize reporter: %w", err)
		}

		// Check if --list-tests flag was provided
		if listTests {
			return runSpecificTests("")
		}

		// Check if --test flag was provided
		if cmd.Flags().Changed("test") {
			return runSpecificTests(testFilter)
		}

		return runAllLevel1Tests()
	},
}

func init() {
	rootCmd.AddCommand(level1Cmd)
	level1Cmd.Flags().StringVar(&testFilter, "test", "", "comma-separated list of specific tests to run (use --test=\"\" to list available tests)")
	level1Cmd.Flags().BoolVar(&listTests, "list-tests", false, "list all available tests")
}

func runAllLevel1Tests() error {
	logger.Info("Running all Level 1 tests")
	rep := reporter.GetReporter()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"gpu_count_check", level1_tests.RunGPUCountCheck},
		{"pcie_error_check", level1_tests.RunPCIeErrorCheck},
		{"rdma_nics_count", level1_tests.RunRDMANicsCount},
		{"rx_discards_check", level1_tests.RunRXDiscardsCheck},
		{"gid_index_check", level1_tests.RunGIDIndexCheck},
	}

	var failedTests []string

	for _, test := range tests {
		logger.Info(fmt.Sprintf("Running test: %s", test.name))
		if err := test.fn(); err != nil {
			logger.Error(fmt.Sprintf("Test %s failed: %v", test.name, err))
			failedTests = append(failedTests, test.name)
		}
	}

	// Get output format from configuration
	outputFormat := viper.GetString("output")
	if outputFormat == "" {
		outputFormat = "table" // Default to table format
	}

	// Generate and write the report with the specified format
	if err := rep.WriteReportWithFormat(outputFormat); err != nil {
		logger.Errorf("Failed to write report: %v", err)
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Print summary only if not using friendly or json format (which should have clean output)
	if outputFormat != "friendly" && outputFormat != "json" {
		rep.PrintSummary()
	}

	if len(failedTests) > 0 {
		logger.Error(fmt.Sprintf("Level 1 tests completed with %d failures: %v", len(failedTests), failedTests))
		// Don't print additional failure messages for JSON or friendly format (keep output clean)
		if outputFormat != "json" && outputFormat != "friendly" {
			fmt.Printf("\n❌ Level 1 diagnostic tests failed: %d out of %d tests failed\n", len(failedTests), len(tests))
			fmt.Printf("Failed tests: %s\n", strings.Join(failedTests, ", "))
		}
		return fmt.Errorf("diagnostic tests failed")
	}

	logger.Info("All Level 1 tests completed successfully")
	// Don't print additional success messages for JSON or friendly format (keep output clean)
	if outputFormat != "json" && outputFormat != "friendly" {
		fmt.Println("\n✅ All Level 1 diagnostic tests passed successfully!")
	}
	return nil
}

func runSpecificTests(testFilter string) error {
	rep := reporter.GetReporter()

	availableTests := []struct {
		name        string
		description string
		fn          func() error
	}{
		{"gpu_count_check", "Check GPU count using nvidia-smi", level1_tests.RunGPUCountCheck},
		{"pcie_error_check", "Check for PCIe errors in system logs", level1_tests.RunPCIeErrorCheck},
		{"rdma_nics_count", "Check RDMA NICs count", level1_tests.RunRDMANicsCount},
		{"rx_discards_check", "Check Network Interface for rx discard", level1_tests.RunRXDiscardsCheck},
		{"gid_index_check", "Check device GID Index are in range ", level1_tests.RunGIDIndexCheck},
	}

	// If testFilter is empty, show available tests
	if testFilter == "" {
		logger.Info("Available Level 1 tests:")
		for _, test := range availableTests {
			logger.Info(fmt.Sprintf("  - %s: %s", test.name, test.description))
		}
		return nil
	}

	// Create test map for lookup
	testMap := make(map[string]func() error)
	for _, test := range availableTests {
		testMap[test.name] = test.fn
	}

	testNames := strings.Split(testFilter, ",")
	logger.Info(fmt.Sprintf("Running specific tests: %v", testNames))

	var failedTests []string

	for _, testName := range testNames {
		testName = strings.TrimSpace(testName)
		if testFn, exists := testMap[testName]; exists {
			logger.Info(fmt.Sprintf("Running test: %s", testName))
			if err := testFn(); err != nil {
				logger.Error(fmt.Sprintf("Test %s failed: %v", testName, err))
				failedTests = append(failedTests, testName)
			}
		} else {
			fmt.Printf("❌ Unknown test: %s\n\n", testName)
			fmt.Printf("Available Level 1 tests:\n")
			for _, test := range availableTests {
				fmt.Printf("  - %s: %s\n", test.name, test.description)
			}
			fmt.Printf("\nUsage examples:\n")
			fmt.Printf("  oci-dr-hpc level1 --test=gpu_count_check\n")
			fmt.Printf("  oci-dr-hpc level1 --test=gpu_count_check,rdma_nics_count\n")
			fmt.Printf("  oci-dr-hpc level1 --list-tests\n")
			return fmt.Errorf("unknown test: %s", testName)
		}
	}

	// Get output format from configuration
	outputFormat := viper.GetString("output")
	if outputFormat == "" {
		outputFormat = "table" // Default to table format
	}

	// Generate and write the report with the specified format
	if err := rep.WriteReportWithFormat(outputFormat); err != nil {
		logger.Errorf("Failed to write report: %v", err)
		return fmt.Errorf("failed to write report: %w", err)
	}

	// Print summary only if not using friendly or json format (which should have clean output)
	if outputFormat != "friendly" && outputFormat != "json" {
		rep.PrintSummary()
	}

	if len(failedTests) > 0 {
		logger.Error(fmt.Sprintf("Selected Level 1 tests completed with %d failures: %v", len(failedTests), failedTests))
		// Don't print additional failure messages for JSON or friendly format (keep output clean)
		if outputFormat != "json" && outputFormat != "friendly" {
			fmt.Printf("\n❌ Level 1 diagnostic tests failed: %d out of %d tests failed\n", len(failedTests), len(testNames))
			fmt.Printf("Failed tests: %s\n", strings.Join(failedTests, ", "))
		}
		return fmt.Errorf("diagnostic tests failed")
	}

	logger.Info("Selected Level 1 tests completed successfully")
	// Don't print additional success messages for JSON or friendly format (keep output clean)
	if outputFormat != "json" && outputFormat != "friendly" {
		fmt.Println("\n✅ All selected Level 1 diagnostic tests passed successfully!")
	}
	return nil
}
