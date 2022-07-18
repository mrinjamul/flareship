package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/models"
	"github.com/mrinjamul/flareship/utils"
	"github.com/spf13/cobra"
)

var (
	flagCheck bool
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "format the records",
	Run: func(cmd *cobra.Command, args []string) {

		if flagRecords == "" {
			flagRecords = "records.json"
		}
		if flagRestricted == "" {
			flagRestricted = "restricted.json"
		}

		// Set domain name if flag exists
		if flagDomain != "" {
			Domain = flagDomain
		}

		if flagCheck {
			var warn bool
			var hasError bool
			var errorsList []string
			var records []models.Records
			records, err := utils.GetRecords(flagRecords)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - cannot able to parse records")
				fmt.Printf("FAIL\t%v\n", err)
				os.Exit(1)
			}
			for id, record := range records {
				fmt.Printf("INFO - id: %d\n", id+1)
				fmt.Printf("INFO - %s: %s %s\n", record.Record.Type, record.Record.Name, record.Record.Content)
				if !record.Record.Proxied && (record.Record.Type == "A" || record.Record.Type == "AAAA" || record.Record.Type == "CNAME") {
					warn = true
					fmt.Println("WARN - Proxied is false")
					fmt.Println("WARN - Please check the record")
				}
				if record.Record.Type == "" {
					fmt.Println("FAIL\trecord type cannot be empty")
					os.Exit(1)
				}
				if record.Record.Name == "" {
					fmt.Println("FAIL\trecord name cannot be empty")
					os.Exit(1)
				}
				if record.Record.Content == "" {
					fmt.Println("FAIL\trecord content cannot be empty")
					os.Exit(1)
				}
			}

			// Check if the records includes restricted subdomains
			var enabledRecordType []string = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
			localRecords, err := utils.GetDNSRecords(flagRecords, enabledRecordType)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - cannot able to parse dns records")
				fmt.Printf("FAIL\t%v\n", err)
				os.Exit(1)
			}
			_, restrictedRecords := utils.RemoveRestrictedSubdomains(flagRestricted, localRecords)
			if len(restrictedRecords) > 0 {
				hasError = true
				fmt.Println("ERROR - Restricted subdomains found")
				fmt.Println("ERROR - Please check the record")
				errorsList = append(errorsList, "Restricted subdomains found")
				fmt.Println()
				// print restricted records
				for _, record := range restrictedRecords {
					fmt.Printf("ERROR - %s: %s %s\n", record.Type, record.Name, record.Content)
				}
			}

			if hasError {
				for _, error := range errorsList {
					fmt.Printf("FAIL\t%s\n", error)
				}
				fmt.Println("TEST\t failed")
				fmt.Println("run `flareship fmt` to fix the errors")
				os.Exit(1)
			}

			fmt.Println()
			fmt.Printf("INFO - %d record(s) found and are valid\n", len(records))
			if warn {
				fmt.Println("WARN - There is some records with warning")
				fmt.Println("WARN - Please check the records")
			}
			fmt.Println("PASS\tok")
			return
		}

		records, err := utils.GetRecords(flagRecords)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR - fail to parse local DNS records")
			os.Exit(1)
		}
		restrictedList := utils.ReadRestrictedRecords(flagRestricted)
		var removeList []int

		var count uint
		var removed bool
		for i := range records {
			var flag bool
			records[i].Record.Proxiable = true
			// Set Proxied to true if the record type is A, AAAA or CNAME
			if (records[i].Record.Type == "A" || records[i].Record.Type == "AAAA" || records[i].Record.Type == "CNAME") && !records[i].Record.Proxied {
				fmt.Println("INFO - Setting Proxied to true")
				records[i].Record.Proxied = true
				count++
				flag = true
			}
			// Set TTL to 1 if the record type is A, AAAA or CNAME
			if (records[i].Record.Type == "A" || records[i].Record.Type == "AAAA" || records[i].Record.Type == "CNAME") && records[i].Record.TTL == 0 {
				fmt.Println("INFO - Setting TTL to auto")
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
			fmt.Println(err)
			fmt.Println("ERROR - fail to convert records")
			os.Exit(1)
		}
		err = os.WriteFile(flagRecords, data, 0644)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR - fail to write records")
			os.Exit(1)
		}
		if removed {
			fmt.Printf("INFO - %d record(s) removed\n", len(removeList))
		}
		fmt.Printf("INFO - %d record(s) formatted\n", count)
		fmt.Println("INFO - fomatting record complete!")
	},
}

func init() {
	fmtCmd.Flags().BoolVarP(&flagCheck, "check", "c", false, "checks if the records has for errors")
	fmtCmd.Flags().StringVarP(&flagRecords, "file", "f", "", "specify the records file")
	fmtCmd.Flags().StringVarP(&flagRestricted, "restricted", "r", "", "specify the restricted domain")
	fmtCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}

// removeRecords removes the records from the records file
func removeRecords(records []models.Records, i int) []models.Records {
	records[i] = records[len(records)-1]
	return records[:len(records)-1]
}
