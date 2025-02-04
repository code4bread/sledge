package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/code4bread/sledge/logger"
)

// log is a logger instance
var log = logger.Logger

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "sledge",
		Short: "CLI to manage GCP Cloud SQL operations",
		Long:  `A demonstration CLI built with Cobra, Viper, and GCP's Cloud SQL Admin API.`,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}


func init() {
	// Global --config flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.sledge.yaml)")
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(UpgradeCmd)
	rootCmd.AddCommand(MigrateCmd)
	rootCmd.AddCommand(DeleteCmd)
	rootCmd.AddCommand(BackupCmd)
    rootCmd.AddCommand(RestoreCmd)
	rootCmd.AddCommand(describeCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".sledge")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
