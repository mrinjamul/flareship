package main

import (
	"encoding/json"

	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/log" // Import the new log package
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

		log.Info("flareship CLI is running 🌟")
		log.Info("sync started...")

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

			log.Info("sync for %s ...", domainName)

			// gather from remote
			log.Info("gathering DNS Records from cloudflare api...")
			registeredRecords := cloudflare.ReadAllRecords(zoneID, token, EnabledRecordType)
			log.Info("got %d registered DNS Records on cf", len(registeredRecords))
			// gather from local
			log.Info("gathering DNS Records from repository...")
			localRecords, err := utils.GetDNSRecords(recordsFile, EnabledRecordType)
			if err != nil {
				log.Error("fail to parse local DNS records: %v", err)
			}
			for id := range localRecords {
				localRecords[id].TTL = 1
				if localRecords[id].Name == "@" {
					localRecords[id].Name = domainName
				} else {
					localRecords[id].Name = localRecords[id].Name + "." + domainName
				}
			}
			log.Info("got %d local CNAME Records in repo", len(localRecords))

			// remove restricted subdomains
			log.Info("removing restricted subdomains...")
			localRecords, removedRecords := utils.RemoveRestrictedSubdomains(restrictedFile, localRecords)
			log.Info("got %d local CNAME Records after removing restricted subdomains", len(localRecords))
			log.Info("removed %d restricted subdomains", len(removedRecords))

			var createdRecords []schema.Record
			var updatedRecords []schema.Record

			log.Info("inspecting DNS records ..")

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
			log.Info("found %d DNS Records to create", len(createdRecords))
			log.Info("found %d DNS Records to update", len(updatedRecords))

			// Create records from the list
			if len(createdRecords) > 0 {
				log.Info("Creating DNS Record(s):")
				for _, r := range createdRecords {
					postBody, err := json.Marshal(r)
					if err != nil {
						log.Error("fail to marshal record while creating: %v", err)
					}
					if !flagDryRun {
						newRecords := cloudflare.CreateRecord(zoneID, token, postBody)
						r = newRecords
					}
					log.Info("+ %-10s %-30s %-40s", r.Type, r.Name, r.Content)
				}
			}
			// Update records from the list
			if len(updatedRecords) > 0 {
				log.Info("Updating DNS Record(s):")
				for _, newRecord := range updatedRecords {
					// Find the old record from registeredRecords
					var oldRecord schema.Record
					for _, regRec := range registeredRecords {
						if regRec.Name == newRecord.Name && regRec.Type == newRecord.Type {
							oldRecord = regRec
							break
						}
					}

					log.Info("~ %-10s %-30s", newRecord.Type, newRecord.Name)
					if oldRecord.Content != newRecord.Content {
						log.Info("- %-40s", oldRecord.Content)
						log.Info("+ %-40s", newRecord.Content)
					}
					if oldRecord.Proxied != newRecord.Proxied {
						log.Info("- Proxied: %t", oldRecord.Proxied)
						log.Info("+ Proxied: %t", newRecord.Proxied)
					}

					postBody, err := json.Marshal(newRecord)
					if err != nil {
						log.Error("fail to marshal record while updating: %v", err)
					}
					if !flagDryRun {
						cloudflare.UpdateRecord(zoneID, token, newRecord.ID, postBody)
					}
				}
			}
			// check for unused records
			log.Info("checking for deleted DNS records...") // Replaced fmt.Println
			var deletedRecords []schema.Record
			// check record which is not in the registeredRecords
			for _, r := range registeredRecords {
				if !utils.RecordContain(localRecords, r) {
					deletedRecords = append(deletedRecords, r)
				}
			}
			log.Info("found %d DNS Records to be delete", len(deletedRecords)) // Replaced fmt.Printf
			// Delete unsed records
			if len(deletedRecords) != 0 {
				log.Info("Deleting DNS Record:") // Replaced fmt.Println
				for _, r := range deletedRecords {
					var result schema.DelResponse
					if !flagDryRun {
						result = cloudflare.DeleteRecord(zoneID, token, r.ID)
						if result.Result.ID == "" {
							log.Error("failed to delete %s:%s", r.Type, r.Name) // Replaced fmt.Println and os.Exit(1)
						}
					}
					log.Info("- %-10s %-30s %-40s", r.Type, r.Name, r.Content)
				}
			} else {
				log.Info("found none") // Replaced fmt.Println
			}
			log.Info("STATUS - %d record(s) created, %d record(s) updated, %d record(s) deleted", len(createdRecords), len(updatedRecords), len(deletedRecords)) // Replaced fmt.Printf
			
			log.Info("sync completed for %s 🎉", domainName) // Replaced fmt.Printf
		}

	},
}

func init() {
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "dry run the sync")
	syncCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}