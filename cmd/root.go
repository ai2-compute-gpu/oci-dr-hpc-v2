// =============================================================================
// cmd/root.go - Main CLI entry point
package cmd

import (
	"fmt"
	"os"

	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	verbose      bool
	outputFormat string
	testLevel    string
)

var rootCmd = &cobra.Command{
	Use:   "oci-dr-hpc",
	Short: "Oracle Cloud Infrastructure Diagnostic and Repair for HPC",
	Long:  `A comprehensive diagnostic and repair tool for HPC environments with GPU and RDMA support.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oci-dr-hpc.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (json|table|friendly)")
	rootCmd.PersistentFlags().StringVarP(&testLevel, "level", "l", "L1", "test level (L1|L2|L3)")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("level", rootCmd.PersistentFlags().Lookup("level"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".oci-dr-hpc")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	}
}
