package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mrinjamul/flareship/internal/config"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	flagConfig string = ""
	AppConfig  *schema.AppConfig
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
	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "specify config file location")

	// PreRun
	_, present := os.LookupEnv("FLARESHIP_CONFIG")
	if present {
		flagConfig = os.Getenv("FLARESHIP_CONFIG")
	}

	var err error
	AppConfig, err = config.LoadConfig(flagConfig)

	if err != nil {
		log.Fatalf("Failed to load config from %s: %v\n", flagConfig, err)
	}

	// fmt.Println(AppConfig)

	err = rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
