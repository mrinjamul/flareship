package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/internal/log"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	flagCheck bool
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "format the records",
	Run: func(cmd *cobra.Command, args []string) {
		for _, domain := range AppConfig.Domains {

			if flagDomain != "" {
				if flagDomain != domain.Name {
					continue
				}
			}

			recordsFile := domain.RecordFile
			restrictedFile := domain.RestrictedFile

			log.Info("Checking records for %s...", domain.Name)

			if recordsFile == "" {
				log.Info("Records file does not exist!")
				continue
			}
			if restrictedFile == "" {
				log.Info("Restricted file does not exist!")
				continue
			}

			if flagCheck {
				var warn bool
				var hasError bool
				var errorsList []string
				var records []schema.Records
				records, err := utils.GetRecords(recordsFile)
				if err != nil {
					log.Error("Failed to parse records: %v", err)
				}
				for id, record := range records {
					log.Info("ID: %d", id+1)
					log.Info("%s: %s %s", record.Record.Type, record.Record.Name, record.Record.Content)
					if !record.Record.Proxied && (record.Record.Type == "A" || record.Record.Type == "AAAA" || record.Record.Type == "CNAME") {
						warn = true
						log.Info("WARN - Proxied is false")
						log.Info("WARN - Please check the record")
					}
					if record.Record.Type == "" {
						log.Error("Record type cannot be empty")
						os.Exit(1)
					}
					if record.Record.Name == "" {
						log.Error("Record name cannot be empty")
						os.Exit(1)
					}
					if record.Record.Content == "" {
						log.Error("Record content cannot be empty")
						os.Exit(1)
					}
				}

				// Check if the records includes restricted subdomains
				var enabledRecordType []string = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
				localRecords, err := utils.GetDNSRecords(recordsFile, enabledRecordType)
				if err != nil {
					log.Error("Failed to parse DNS records: %v", err)
				}
				_, restrictedRecords := utils.RemoveRestrictedSubdomains(restrictedFile, localRecords)
				if len(restrictedRecords) > 0 {
					hasError = true
					log.Info("Restricted subdomains found")
					log.Info("Please check the record")
					errorsList = append(errorsList, "Restricted subdomains found")
					fmt.Println()
					// print restricted records
					for _, record := range restrictedRecords {
						log.Info("%s: %s %s", record.Type, record.Name, record.Content)
					}
				}

				if hasError {
					for _, error := range errorsList {
						log.Info("%s", error)
					}
					log.Info("Run `flareship fmt` to fix the errors")
					log.Error("Test failed")
				}

				log.Info("%d record(s) found and are valid", len(records))
				if warn {
					log.Info("WARN - Some records have warnings")
					log.Info("WARN - Please check the records")
				}
				log.Info("PASS - All checks passed.")
				continue
			}

			records, err := utils.GetRecords(recordsFile)
			if err != nil {
				log.Error("Failed to parse local DNS records: %v", err)
			}
			restrictedList := utils.ReadRestrictedRecords(restrictedFile)
			var removeList []int

			var count uint
			var removed bool
			for i := range records {
				var flag bool
				records[i].Record.Proxiable = true
				// Set Proxied to true if the record type is A, AAAA or CNAME
				if (records[i].Record.Type == "A" || records[i].Record.Type == "AAAA" || records[i].Record.Type == "CNAME") && !records[i].Record.Proxied {
					log.Info("Setting Proxied to true for %s", records[i].Record.Name)
					records[i].Record.Proxied = true
					count++
					flag = true
				}
				// Set TTL to 1 if the record type is A, AAAA or CNAME
				if (records[i].Record.Type == "A" || records[i].Record.Type == "AAAA" || records[i].Record.Type == "CNAME") && records[i].Record.TTL == 0 {
					log.Info("Setting TTL to auto for %s", records[i].Record.Name)
					records[i].Record.TTL = 1
					if !flag {
						count++
					}
					flag = true
				}
				if utils.IsRestricted(records[i].Record.Name, restrictedList) {
					// remove this record from the records
					removeList = append(removeList, i)
				}

			}
			// remove restricted records
			if len(removeList) > 0 {
				if ok := utils.ConfirmPrompt("Do you want to remove restricted subdomains?"); ok {
					count += uint(len(removeList))
					removed = true
					for _, i := range removeList {
						records = removeRecords(records, i)
					}
				}
			}
			// write the records to the file
			data, err := json.MarshalIndent(records, "", "\t")
			if err != nil {
				log.Error("Failed to convert records to JSON: %v", err)
			}
			err = os.WriteFile(recordsFile, data, 0644)
			if err != nil {
				log.Error("Failed to write records to file: %v", err)
			}
			if removed {
				log.Info("%d record(s) removed", len(removeList))
			}
			log.Info("%d record(s) formatted", count)
			log.Info("Formatting record complete!")
		}
	},
}

func init() {
	fmtCmd.Flags().BoolVarP(&flagCheck, "check", "c", false, "checks if the records has for errors")
	fmtCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}

// removeRecords removes the records from the records file
func removeRecords(records []schema.Records, i int) []schema.Records {
	records[i] = records[len(records)-1]
	return records[:len(records)-1]
}
