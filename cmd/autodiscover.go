package cmd

import (
	"fmt"

	"github.com/oracle/oci-dr-hpc-v2/internal/autodiscover"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/spf13/cobra"
)

var autodiscoverCmd = &cobra.Command{
	Use:   "autodiscover",
	Short: "Generate a logical model of the GPU Hardware",
	Long: `Autodiscover generates a logical model of the GPU Hardware by analyzing the system configuration,
detecting GPU devices, RDMA NICs, and other HPC components to create a comprehensive hardware inventory.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Starting autodiscover process")

		// Call the autodiscover functionality
		autodiscover.Run()

		logger.Info("Autodiscover completed successfully")
		fmt.Println("✅ Hardware autodiscovery completed successfully!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(autodiscoverCmd)
}
