package cmd

import (
	"fmt"

	custom_script "github.com/oracle/oci-dr-hpc-v2/internal/custom-script"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scriptPath          string
	limitsFile          string
	recommendationsFile string
)

var customScriptCmd = &cobra.Command{
	Use:   "custom-script",
	Short: "Execute custom diagnostic scripts with configuration support",
	Long: `Execute custom diagnostic scripts for HPC environments with support for test limits and recommendations configuration.
The custom-script command allows you to run custom Python or shell scripts while leveraging the application's
configuration system for test limits and recommendations.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting custom script execution")

		// Validate required flags
		if scriptPath == "" {
			return fmt.Errorf("script path is required. Use --script flag to specify the script to execute")
		}

		// Get output format from configuration
		outputFormat := viper.GetString("output")
		if outputFormat == "" {
			outputFormat = "table" // Default to table format
		}

		// Execute the custom script with configuration
		if err := custom_script.ExecuteScript(scriptPath, limitsFile, recommendationsFile, outputFormat); err != nil {
			logger.Errorf("Failed to execute custom script: %v", err)
			return fmt.Errorf("failed to execute custom script: %w", err)
		}

		logger.Info("Custom script execution completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(customScriptCmd)
	customScriptCmd.Flags().StringVarP(&scriptPath, "script", "s", "", "path to custom script to execute (required)")
	customScriptCmd.Flags().StringVar(&limitsFile, "limits-file", "", "path to test limits configuration file")
	customScriptCmd.Flags().StringVar(&recommendationsFile, "recommendations-file", "", "path to recommendations configuration file")
	customScriptCmd.MarkFlagRequired("script")
}
