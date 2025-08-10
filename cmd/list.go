package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/spf13/cobra"
)

var (
	flagLocal bool
	flagTypes string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all records from remote/local",
	Run: func(cmd *cobra.Command, args []string) {
		// all type of dns records
		types := []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
		if flagTypes != "" {
			types = strings.Split(flagTypes, ",")
		}

		for _, domain := range AppConfig.Domains {

			recordFile := domain.RecordFile
			Domain := domain.Name
			if len(EnabledRecordType) == 0 {
				EnabledRecordType = domain.RecordTypes
			}
			// list records from local json file
			if flagLocal {
				fmt.Println("INFO - gathering DNS Records from local ...")
				localRecords, err := utils.GetDNSRecords(recordFile, EnabledRecordType)
				if err != nil {
					fmt.Println(err)
					fmt.Println("ERROR - fail to parse local DNS records")
					os.Exit(1)
				}
				for _, record := range localRecords {
					fmt.Printf("%s: %s.%s -> %s\t%d\n", record.Type, record.Name, Domain, record.Content, record.TTL)
				}
				fmt.Printf("INFO - got %d registered DNS Records from local records \n", len(localRecords))
				continue
			}

			// gather from remote
			fmt.Printf("INFO - gathering DNS Records for %s from cloudflare api...\n", Domain)
			allRecords := cloudflare.ReadAllRecords(domain.ZoneID, domain.CFToken, types)
			for _, record := range allRecords {
				fmt.Printf("%s: %s.%s -> %s\t%d\n", record.Type, record.Name, Domain, record.Content, record.TTL)
			}
			fmt.Printf("INFO - got %d registered DNS Records on cloudflare for %s \n", len(allRecords), Domain)
		}

	},
}

func init() {
	listCmd.Flags().StringVarP(&flagTypes, "type", "t", "", "specify the types of records")
	listCmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "specify the target to list e.g. local")
}
