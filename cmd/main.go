package main

import (
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/utils"
	"github.com/spf13/cobra"
)

var (
	flagConfig     string = ""
	flagRecords    string
	flagRestricted string
	ZoneID         string
	CFToken        string
)

func init() {
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mrinjamul",
		Short: "mrinjamul.in CLI",
		Run: func(cmd *cobra.Command, args []string) {
			// Root command
			var tip string = "tip: "
			tip += utils.GenTips()
			fmt.Println(tip)
		},
	}

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(fmtCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(listCmd)
	// add flags
	// rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "config file")

	// PreRun
	_, present := os.LookupEnv("CONFIG_FILE")
	if present {
		flagConfig = os.Getenv("CONFIG_FILE")
	}
	// get config variables
	flagDomain, flagRecords, flagRestricted, CFToken, ZoneID, EnabledRecordType = utils.GetConfig(flagConfig)

	// get records file
	_, present = os.LookupEnv("RECORD_FILE")
	if present {
		flagRecords = os.Getenv("RECORD_FILE")
	}
	// get restricted file
	_, present = os.LookupEnv("RESTRICTED_FILE")
	if present {
		flagRestricted = os.Getenv("RESTRICTED_FILE")
	}
	// Get domain Name
	_, present = os.LookupEnv("DOMAIN_NAME")
	if present {
		flagDomain = os.Getenv("DOMAIN_NAME")
	}
	// get CF TOKEN
	_, present = os.LookupEnv("CF_TOK")
	if present {
		CFToken = os.Getenv("CF_TOK")
	}
	// get CF ZONE ID
	_, present = os.LookupEnv("CF_ZID")
	if present {
		ZoneID = os.Getenv("CF_ZID")
	}

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
