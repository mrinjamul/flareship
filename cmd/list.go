package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrinjamul/flareship/utils"
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

		// list records from local json file
		if flagLocal {
			if flagRecords == "" {
				flagRecords = "records.json"
			}
			fmt.Println("INFO - gathering DNS Records from local ...")
			localRecords, err := utils.GetDNSRecords(flagRecords, EnabledRecordType)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - fail to parse local DNS records")
				os.Exit(1)
			}
			for _, record := range localRecords {
				fmt.Printf("%s: %s.%s -> %s\t%d\n", record.Type, record.Name, Domain, record.Content, record.TTL)
			}
			fmt.Printf("INFO - got %d registered DNS Records on cf \n", len(localRecords))
			return
		}

		// Set domain name if flag exists
		if flagDomain != "" {
			Domain = flagDomain
		}
		// gather from remote
		fmt.Println("INFO - gathering DNS Records from cloudflare api...")
		allRecords := GetRecords(types)
		for _, record := range allRecords {
			fmt.Printf("%s: %s.%s -> %s\t%d\n", record.Type, record.Name, Domain, record.Content, record.TTL)
		}
		fmt.Printf("INFO - got %d registered DNS Records on cf \n", len(allRecords))
	},
}

func init() {
	listCmd.Flags().StringVarP(&flagTypes, "type", "t", "", "specify the types of records")
	listCmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "specify the target to list e.g. local")
}
