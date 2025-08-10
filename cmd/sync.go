package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	flagDryRun bool
)

var (
	// EnabledRecordType []string = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
	// EnabledRecordType specifies the record types that will be synced
	EnabledRecordType []string = []string{}
)

// Sync sync the records
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync with remote DNS.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("flareship CLI is running 🌟")
		fmt.Println("sync started...")

		for _, domain := range AppConfig.Domains {
			if flagDomain != "" {
				if flagDomain != domain.Name {
					continue
				}
			}
			// Set domain name if flag exist
			domainName := domain.Name
			recordsFile := domain.RecordFile
			restrictedFile := domain.RestrictedFile
			zoneID := domain.ZoneID
			token := domain.CFToken

			// Set enabled records if it is null
			if len(domain.RecordTypes) == 0 {
				EnabledRecordType = []string{"A", "CNAME"}
			} else {
				EnabledRecordType = domain.RecordTypes
			}

			fmt.Printf("sync for %s ...\n", domainName)

			// gather from remote
			fmt.Println("INFO - gathering DNS Records from cloudflare api...")
			registeredRecords := cloudflare.ReadAllRecords(zoneID, token, EnabledRecordType)
			fmt.Printf("INFO - got %d registered DNS Records on cf \n", len(registeredRecords))
			// gather from local
			fmt.Println("INFO - gathering DNS Records from repository...")
			localRecords, err := utils.GetDNSRecords(recordsFile, EnabledRecordType)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - fail to parse local DNS records")
				os.Exit(1)
			}
			for id := range localRecords {
				localRecords[id].TTL = 1
				if localRecords[id].Name == "@" {
					localRecords[id].Name = domainName
				} else {
					localRecords[id].Name = localRecords[id].Name + "." + domainName
				}
			}
			fmt.Printf("INFO - got %d local CNAME Records in repo \n", len(localRecords))

			// remove restricted subdomains
			fmt.Println("INFO - removing restricted subdomains...")
			localRecords, removedRecords := utils.RemoveRestrictedSubdomains(restrictedFile, localRecords)
			fmt.Printf("INFO - got %d local CNAME Records after removing restricted subdomains \n", len(localRecords))
			fmt.Printf("INFO - removed %d restricted subdomains \n", len(removedRecords))

			var createdRecords []schema.Record
			var updatedRecords []schema.Record

			fmt.Println("INFO - inspecting DNS records ..")

			for _, record := range localRecords {
				r := utils.FindRecordByName(registeredRecords, record.Name)
				if r.ID != "" {
					if r.Content != record.Content || r.Proxied != record.Proxied || r.Name != record.Name {
						record.ID = r.ID
						updatedRecords = append(updatedRecords, record)
					}
				} else {
					createdRecords = append(createdRecords, record)
				}
			}
			fmt.Printf("INFO - found %d DNS Records to create \n", len(createdRecords))
			fmt.Printf("INFO - found %d DNS Records to update \n", len(updatedRecords))

			// Create records from the list
			if len(createdRecords) > 0 {
				fmt.Println(" INFO - Creating DNS Record(s):")
				for _, r := range createdRecords {
					postBody, err := json.Marshal(r)
					if err != nil {
						fmt.Println(err)
						fmt.Println("ERROR - fail to marshal record while creating")
						os.Exit(1)
					}
					if !flagDryRun {
						newRecords := cloudflare.CreateRecord(zoneID, token, postBody)
						r = newRecords
					}
					fmt.Printf("%s %s: %s %s\n", r.ID, r.Type, r.Name, r.Content)
				}
			}
			// Update records from the list
			if len(updatedRecords) > 0 {
				fmt.Println("INFO - Updating DNS Record(s):")
				for _, r := range updatedRecords {
					fmt.Println(r)
					postBody, err := json.Marshal(r)
					if err != nil {
						fmt.Println(err)
						fmt.Println("ERROR - fail to marshal record while updating")
						os.Exit(1)
					}
					if !flagDryRun {
						r = cloudflare.UpdateRecord(zoneID, token, r.ID, postBody)
					}
					fmt.Printf("%s %s: %s %s\n", r.ID, r.Type, r.Name, r.Content)
				}
			}
			// check for unused records
			fmt.Println("INFO - checking for deleted DNS records...")
			var deletedRecords []schema.Record
			// check record which is not in the registeredRecords
			for _, r := range registeredRecords {
				if !utils.RecordContain(localRecords, r) {
					deletedRecords = append(deletedRecords, r)
				}
			}
			fmt.Printf("INFO - found %d DNS Records to be delete \n", len(deletedRecords))
			// Delete unsed records
			if len(deletedRecords) != 0 {
				fmt.Println("Deleting DNS Record:")
				for _, r := range deletedRecords {
					var result schema.DelResponse
					if !flagDryRun {
						result = cloudflare.DeleteRecord(zoneID, token, r.ID)
						if result.Result.ID == "" {
							fmt.Println("ERROR - failed to delete " + r.Type + ":" + r.Name)
						}
					}
					fmt.Printf("%s: %s %s\n", result.Result.ID, r.Name, r.Content)
				}
			} else {
				fmt.Println("INFO - found none")
			}
			fmt.Printf("STATUS - %d record(s) created, %d record(s) updated, %d record(s) deleted\n", len(createdRecords), len(updatedRecords), len(deletedRecords))
			fmt.Println("")
			fmt.Printf("sync completed for %s 🎉\n", domainName)
		}

	},
}

func init() {
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "dry run the sync")
	syncCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}
