package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	flagDomain string
)

// versionCmd represents the version command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup DNS records to file.",
	Run: func(cmd *cobra.Command, args []string) {
		for _, domain := range AppConfig.Domains {
			if flagDomain != "" {
				if flagDomain != domain.Name {
					continue
				}
			}
			var records []schema.Records
			var cfrecords []schema.Record

			if len(EnabledRecordType) == 0 {
				EnabledRecordType = domain.RecordTypes
			} else {
				EnabledRecordType = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
			}

			if flagTypes != "" {
				if flagTypes == "all" {
					EnabledRecordType = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
				} else {

					types := strings.Split(flagTypes, ",")
					for i := range types {
						types[i] = strings.ToUpper(strings.TrimSpace(types[i]))
					}
					EnabledRecordType = types
				}
			}

			fmt.Println("INFO - backup started...")
			cfrecords = cloudflare.ReadAllRecords(domain.ZoneID, domain.CFToken, EnabledRecordType)
			for _, record := range cfrecords {
				var r schema.Records
				suffix := "." + domain.Name
				record.Name = strings.TrimSuffix(record.Name, suffix)
				r.Record = record
				records = append(records, r)
			}
			fmt.Println("INFO - backuping to file...")
			err := backupRecords(records, domain.Name)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - cannot able to backup records")
				fmt.Printf("FAIL\t%v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("INFO - backup completed...")
	},
}

func init() {
	backupCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
	backupCmd.Flags().StringVarP(&flagTypes, "type", "t", "", "specify the types of records")
}

func backupRecords(records []schema.Records, domainName string) error {
	var configFile string

	date := utils.NewDate()
	num := utils.RandomNumber()
	configFile = "dns_records_" + domainName + "_" + date + "_" + num + ".json"
	fmt.Println(configFile)
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}
	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
