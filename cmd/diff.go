package main

import (
	"github.com/mrinjamul/flareship/internal/cloudflare"
	"github.com/mrinjamul/flareship/internal/log"
	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show differences between local and remote DNS records",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("flareship CLI is running 🌟")
		log.Info("diff started...")

		for _, domain := range AppConfig.Domains {
			if flagDomain != "" {
				if flagDomain != domain.Name {
					continue
				}
			}

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

			log.Info("diff for %s ...", domainName)

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
			var deletedRecords []schema.Record

			log.Info("inspecting DNS records for differences..")

			// Find created and updated records
			for _, localRecord := range localRecords {
				found := false
				for _, registeredRecord := range registeredRecords {
					if localRecord.Name == registeredRecord.Name && localRecord.Type == registeredRecord.Type {
						found = true
						if localRecord.Content != registeredRecord.Content || localRecord.Proxied != registeredRecord.Proxied {
							updatedRecords = append(updatedRecords, localRecord)
						}
						break
					}
				}
				if !found {
					createdRecords = append(createdRecords, localRecord)
				}
			}

			// Find deleted records
			for _, registeredRecord := range registeredRecords {
				found := false
				for _, localRecord := range localRecords {
					if registeredRecord.Name == localRecord.Name && registeredRecord.Type == localRecord.Type {
						found = true
						break
					}
				}
				if !found {
					deletedRecords = append(deletedRecords, registeredRecord)
				}
			}

			log.Info("Differences for %s:", domainName)
			log.Info("--------------------------------------------------------------------------------")

			if len(createdRecords) > 0 {
				log.Info("Records to be created:")
				for _, r := range createdRecords {
					log.Info("+ %-10s %-30s %-40s", r.Type, r.Name, r.Content)
				}
			}

			if len(updatedRecords) > 0 {
				log.Info("Records to be updated:")
				for _, r := range updatedRecords {
					// For updated records, show both old and new values if possible
					// This requires fetching the old record's content, which is not directly available here.
					// For simplicity, just show the new value for now.
					log.Info("~ %-10s %-30s %-40s (new)", r.Type, r.Name, r.Content)
				}
			}

			if len(deletedRecords) > 0 {
				log.Info("Records to be deleted:")
				for _, r := range deletedRecords {
					log.Info("- %-10s %-30s %-40s", r.Type, r.Name, r.Content)
				}
			}

			if len(createdRecords) == 0 && len(updatedRecords) == 0 && len(deletedRecords) == 0 {
				log.Info("No differences found.")
			}
			log.Info("--------------------------------------------------------------------------------")
			log.Info("diff completed for %s 🎉", domainName)
		}
	},
}

func init() {
	diffCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}
