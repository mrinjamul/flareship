package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/models"
	"github.com/mrinjamul/flareship/utils"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export DNS records to file.",
	Run: func(cmd *cobra.Command, args []string) {
		var records []models.Records
		var cfrecords []models.Record
		// Set domain name if flag exists
		if flagDomain != "" {
			Domain = flagDomain
		}
		fmt.Println("INFO - export started...")
		cfrecords = GetRecords(EnabledRecordType)
		for _, record := range cfrecords {
			var r models.Records
			r.Record = record
			records = append(records, r)
		}
		fmt.Println("INFO - exporting to file...")
		err := ExportRecords(records)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR - cannot able to export records")
			fmt.Printf("FAIL\t%v\n", err)
			os.Exit(1)
		}
		fmt.Println("INFO - export completed...")
	},
}

func init() {
	exportCmd.Flags().StringVarP(&flagRecords, "file", "f", "", "specify the export file")
	exportCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}

func ExportRecords(records []models.Records) error {
	var configFile string
	if flagRecords != "" {
		configFile = flagRecords
	} else {
		date := utils.NewDate()
		num := utils.RandomNumber()
		configFile = "dns_records_" + date + "_" + num + ".json"
	}
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
