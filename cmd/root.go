// =============================================================================
// cmd/root.go - Main CLI entry point
package cmd

import (
	"fmt"
	"os"

	"github.com/oracle/oci-dr-hpc-v2/internal/config"
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version is set at build time
var version = "dev"

// SetVersion sets the version from main package
func SetVersion(v string) {
	version = v
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

var (
	cfgFile      string
	verbose      bool
	outputFormat string
	testLevel    string
	showVersion  bool
)

var rootCmd = &cobra.Command{
	Use:   "oci-dr-hpc",
	Short: "Oracle Cloud Infrastructure Diagnostic and Repair for HPC",
	Long:  `A comprehensive diagnostic and repair tool for HPC environments with GPU and RDMA support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			fmt.Printf("oci-dr-hpc-v2 version %s\n", GetVersion())
			return nil
		}
		return cmd.Help()
	},
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
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "show version information")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("level", rootCmd.PersistentFlags().Lookup("level"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigType("yaml")
		
		// Add system-wide config path first
		viper.AddConfigPath("/etc")
		
		// Add user config path second (higher priority)
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		
		// Set config name once after all paths are added
		viper.SetConfigName("oci-dr-hpc")
	}

	viper.SetEnvPrefix("OCI_DR_HPC")
	viper.AutomaticEnv()
	
	// Explicitly bind environment variables for nested keys
	viper.BindEnv("logging.level", "OCI_DR_HPC_LOGGING_LEVEL")
	viper.BindEnv("logging.file", "OCI_DR_HPC_LOGGING_FILE")

	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
		logger.Info("DEBUG: Viper logging.file value:", viper.GetString("logging.file"))
		logger.Info("DEBUG: Viper logging.level value:", viper.GetString("logging.level"))
	} else {
		logger.Error("DEBUG: Failed to read viper config:", err)
	}
	
	// Always try to load config (from file or env vars)
	cfg, err := config.LoadConfig()
	logger.Info("DEBUG: Config loading result - err:", err)
	if err == nil {
		logger.Info("DEBUG: Config loaded - logging.file:", cfg.Logging.File, "logging.level:", cfg.Logging.Level)
		
		// Set log level from config (could be from file or env var)
		if cfg.Logging.Level != "" {
			logger.SetLogLevel(cfg.Logging.Level)
		}
		
		// Initialize logger with file if specified
		if cfg.Logging.File != "" {
			logger.Info("DEBUG: Attempting to initialize file logging with path:", cfg.Logging.File)
			if err := logger.InitLoggerWithLevel(cfg.Logging.File, cfg.Logging.Level); err != nil {
				logger.Error("Failed to initialize file logging:", err)
			} else {
				logger.Info("DEBUG: File logging initialized successfully")
			}
		} else {
			logger.Info("DEBUG: No log file specified in config")
		}
	} else {
		logger.Error("DEBUG: Failed to load config:", err)
	}
}
