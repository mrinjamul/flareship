package main

import (
	"fmt" // Keep fmt for Sprintf for table formatting
	"strings"

	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/log" // Import the new log package
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

		for _, domain := range AppConfig.Domains {

			if flagDomain != "" {
				if flagDomain != domain.Name {
					continue
				}
			}

			recordFile := domain.RecordFile
			domainName := domain.Name

			if len(EnabledRecordType) == 0 {
				EnabledRecordType = domain.RecordTypes
			}

			if flagTypes != "" {
				EnabledRecordType = strings.Split(flagTypes, ",")
				if flagTypes == "all" {
					// all type of dns records
					EnabledRecordType = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
				}
			}

			// list records from local json file
			if flagLocal {
				log.Info("gathering DNS Records from local ...")
				localRecords, err := utils.GetDNSRecords(recordFile, EnabledRecordType)
				if err != nil {
					log.Error("fail to parse local DNS records: %v", err)
				}
				log.Info("DNS Records for %s (local):", domainName)
				log.Info("--------------------------------------------------------------------------------")
				log.Info("%-10s %-30s %-40s %-5s", "TYPE", "NAME", "CONTENT", "TTL")
				log.Info("--------------------------------------------------------------------------------")
				for _, record := range localRecords {
					log.Info("%-10s %-30s %-40s %-5d", record.Type, fmt.Sprintf("%s.%s", record.Name, domainName), record.Content, record.TTL)
				}
				log.Info("--------------------------------------------------------------------------------")
				log.Info("got %d registered DNS Records from local records", len(localRecords))
				continue
			}

			// gather from remote
			log.Info("gathering DNS Records for %s from cloudflare api...", domainName)
			allRecords := cloudflare.ReadAllRecords(domain.ZoneID, domain.CFToken, EnabledRecordType)
			log.Info("DNS Records for %s (remote):", domainName)
			log.Info("--------------------------------------------------------------------------------")
			log.Info("%-10s %-30s %-40s %-5s", "TYPE", "NAME", "CONTENT", "TTL")
			log.Info("--------------------------------------------------------------------------------")
			for _, record := range allRecords {
				log.Info("%-10s %-30s %-40s %-5d", record.Type, record.Name, record.Content, record.TTL)
			}
			log.Info("--------------------------------------------------------------------------------")
			log.Info("got %d registered DNS Records on cloudflare for %s", len(allRecords), domainName)
		}

	},
}

func init() {
	listCmd.Flags().StringVarP(&flagTypes, "type", "t", "", "specify the types of records")
	listCmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "specify the target to list e.g. local")
	listCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}
