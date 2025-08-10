package main

import (
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/internal/config"
	"github.com/mrinjamul/flareship/internal/log" // Import the new log package
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	flagConfig  string = ""
	flagVerbose bool   // New global verbose flag
	AppConfig   *schema.AppConfig
)

func init() {
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "flareship",
		Short: "flareship CLI",
		Run: func(cmd *cobra.Command, args []string) {
			// Root command
			var tip string = "tip: "
			tip += utils.GenTips()
			fmt.Println(tip)
		},
	}

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(backupCmd)
	// add flags
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "enable verbose output") // Add verbose flag
	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "specify config file location")

	// Initialize logger verbosity
	log.SetVerbose(flagVerbose)

	// PreRun
	_, present := os.LookupEnv("FLARESHIP_CONFIG")
	if present {
		flagConfig = os.Getenv("FLARESHIP_CONFIG")
	}

	var err error
	AppConfig, err = config.LoadConfig(flagConfig)

	if err != nil {
		log.Error("Failed to load config from %s: %v", flagConfig, err) // Use log.Error
	}

	// fmt.Println(AppConfig)

	err = rootCmd.Execute()
	if err != nil {
		log.Error("%v", err) // Use log.Error
	}
}
