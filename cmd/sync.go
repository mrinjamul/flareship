package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/mrinjamul/flareship/models"
	"github.com/mrinjamul/flareship/utils"
	"github.com/spf13/cobra"
)

var (
	flagDryRun  bool
	flagProxied bool
	flagDomain  string
)

var (
	// BaseAPI is the base url for cloudflare api
	BaseAPI string = "https://api.cloudflare.com/client/v4/"
	// DomainName sets the domain name
	Domain string = "mrinjamul.in"
	// Endpoint specifies the endpoint of the cloudflare api
	Endpoint string
	// EnabledRecordType []string = []string{"A", "AAAA", "CNAME", "TXT", "MX", "SRV"}
	// EnabledRecordType specifies the record types that will be synced
	EnabledRecordType []string = []string{"A", "CNAME"}
)

// Sync sync the records
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync with remote DNS.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("mrinjamul.in CLI is running 🌟")
		fmt.Println("sync started...")

		// Set domain name if flag exists
		if flagDomain != "" {
			Domain = flagDomain
		}
		// Set enabled records if it is null
		if len(EnabledRecordType) == 0 {
			EnabledRecordType = []string{"A", "CNAME"}
		}

		// gather from remote
		fmt.Println("INFO - gathering DNS Records from cloudflare api...")
		registeredRecords := GetRecords(EnabledRecordType)
		fmt.Printf("INFO - got %d registered DNS Records on cf \n", len(registeredRecords))
		// gather from local
		if flagRecords == "" {
			flagRecords = "records.json"
		}
		fmt.Println("INFO - gathering DNS Records from repository...")
		localRecords, err := utils.GetDNSRecords(flagRecords, EnabledRecordType)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ERROR - fail to parse local DNS records")
			os.Exit(1)
		}
		for id := range localRecords {
			if flagProxied {
				// enable always proxied
				localRecords[id].Proxied = true
			}
			localRecords[id].TTL = 1
			if localRecords[id].Name == "@" {
				localRecords[id].Name = Domain
			} else {
				localRecords[id].Name = localRecords[id].Name + "." + Domain
			}
		}
		fmt.Printf("INFO - got %d local CNAME Records in repo \n", len(localRecords))

		// remove restricted subdomains
		if flagRestricted == "" {
			flagRestricted = "restricted.json"
		}
		fmt.Println("INFO - removing restricted subdomains...")
		localRecords, removedRecords := utils.RemoveRestrictedSubdomains(flagRestricted, localRecords)
		fmt.Printf("INFO - got %d local CNAME Records after removing restricted subdomains \n", len(localRecords))
		fmt.Printf("INFO - removed %d restricted subdomains \n", len(removedRecords))

		var createdRecords []models.Record
		var updatedRecords []models.Record

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
					newRecords := CreateRecord(postBody)
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
					r = UpdateRecord(r.ID, postBody)
				}
				fmt.Printf("%s %s: %s %s\n", r.ID, r.Type, r.Name, r.Content)
			}
		}
		// check for unused records
		fmt.Println("INFO - checking for deleted DNS records...")
		var deletedRecords []models.Record
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
				var result models.DelResponse
				if !flagDryRun {
					result = DeleteRecord(r.ID)
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
		fmt.Println("sync completed 🎉")
	},
}

func init() {
	syncCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "dry run the sync")
	syncCmd.Flags().BoolVarP(&flagProxied, "proxied", "p", false, "set all records proxied")
	syncCmd.Flags().StringVarP(&flagRecords, "file", "f", "", "specify the records file")
	syncCmd.Flags().StringVarP(&flagRestricted, "restricted", "r", "", "specify the restricted subdomains file")
	syncCmd.Flags().StringVar(&flagDomain, "domain", "", "specify the domain name")
}

// GetRecords returns all records from cloudflare api
func GetRecords(recordTypes []string) []models.Record {
	query := url.Values{}
	var records []models.Record
	var results []models.Result
	for _, t := range recordTypes {
		query.Add("type", t)
		perPage := 100
		page := 1
		query.Add("per_page", strconv.Itoa(perPage))
		for ok := true; ok; ok = (len(results) == perPage) {
			query.Add("page", strconv.Itoa(page))
			query := query.Encode()
			Endpoint = "zones/" + ZoneID + "/dns_records?" + query
			resp, err := utils.CFFetch(BaseAPI, Endpoint, CFToken)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ERROR - fail to fetch records")
				os.Exit(1)
			}
			if !resp.Success {
				break
			}
			results = resp.Result
			records = utils.Concat(records, results)
		}
		query.Del("type")
	}
	return records
}

// CreateRecord create a new record
func CreateRecord(postBody []byte) models.Record {
	var result models.Result
	Endpoint = "zones/" + ZoneID + "/dns_records"
	resp, err := utils.CFPost("POST", BaseAPI, Endpoint, postBody, CFToken)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to create records")
		os.Exit(1)
	}
	result = resp.Result
	record := utils.ConcatOne(models.Record{}, result)
	return record
}

// UpdateRecord updates a record
func UpdateRecord(recordID string, postBody []byte) models.Record {
	var result models.Result
	Endpoint = "zones/" + ZoneID + "/dns_records/" + recordID
	resp, err := utils.CFPost("PUT", BaseAPI, Endpoint, postBody, CFToken)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to update records")
		os.Exit(1)
	}
	result = resp.Result
	record := utils.ConcatOne(models.Record{}, result)
	return record
}

// DeleteRecord delete a record
func DeleteRecord(recordID string) models.DelResponse {
	Endpoint = "zones/" + ZoneID + "/dns_records/" + recordID
	resp, err := utils.CFDelete(BaseAPI, Endpoint, CFToken)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to delete records")
		os.Exit(1)
	}
	return resp
}
