package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mrinjamul/flareship/pkg/schema"
)

// HomeDir returns the home directory of the current user
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	home, _ := os.UserHomeDir()
	return home
}

// GenTips generates random tips
func GenTips() string {
	tips := []string{
		"Use `flareship init` to generate config file",
		"Use `flareship sync --dry-run` to see what will be synced",
		"Use `flareship sync` to sync your records",
		"Use `flareship sync --domain [url]` to specify the root domain",
		"Use `flareship fmt --check` to check records file",
		"Use `flareship fmt` to format records file",
		"Use `flareship fmt --domain [url]` to specify the root domain",
		"Use `flareship fmt --dry-run` to see what will be formatted",
	}
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	return tips[r.Intn(len(tips))]
}

// GetRecords parse records from records file
func GetRecords(filename string) ([]schema.Records, error) {
	var records []schema.Records
	data, err := os.ReadFile(filename)
	if err != nil {
		return []schema.Records{}, err
	}
	err = json.Unmarshal(data, &records)
	if err != nil {

		return []schema.Records{}, err
	}
	return records, nil
}

// TypeContains checks if a given type is in the given types
func TypeContains(types []string, typeToCheck string) bool {
	for _, t := range types {
		if t == typeToCheck {
			return true
		}
	}
	return false
}

// GetDNSRecords returns the DNS records with the given type
func GetDNSRecords(filename string, enabledRecordType []string) ([]schema.Record, error) {
	var records []schema.Record
	entries, err := GetRecords(filename)
	if err != nil {
		return []schema.Record{}, err
	}
	for _, entry := range entries {
		if TypeContains(enabledRecordType, entry.Record.Type) {
			records = append(records, entry.Record)
		}
	}
	return records, nil
}

// FindRecordByName returns the record from name
func FindRecordByName(records []schema.Record, name string) schema.Record {
	for _, record := range records {
		if record.Name == name {
			return record
		}
	}
	return schema.Record{}
}

// FindRecordID returns the record ID from name
func FindRecordID(records []schema.Record, name string) string {
	for _, r := range records {
		if r.Name == name {
			return r.ID
		}
	}
	return ""
}

// RecordContain checks if a single record is in the records
func RecordContain(records []schema.Record, record schema.Record) bool {
	for _, r := range records {
		if r.Name == record.Name {
			return true
		}
	}
	return false
}

// RecordContains checks if the sub-record is in the records
func RecordContains(records []schema.Record, subrecords []schema.Record) bool {
	for _, r := range subrecords {
		if !RecordContain(records, r) {
			return false
		}
	}
	return true
}

// Concat converts results to records
func Concat(records []schema.Record, result []schema.Result) []schema.Record {
	for _, r := range result {
		// record := make(schema.Record, 0)
		var record schema.Record
		record.ID = r.ID
		record.Type = r.Type
		record.Name = r.Name
		record.Content = r.Content
		record.Proxiable = r.Proxiable
		record.Proxied = r.Proxied
		record.TTL = r.TTL
		records = append(records, record)
	}
	return records
}

// ConcatOne concatenates from the result to record
func ConcatOne(record schema.Record, result schema.Result) schema.Record {
	record.ID = result.ID
	record.Type = result.Type
	record.Name = result.Name
	record.Content = result.Content
	record.Proxiable = result.Proxiable
	record.Proxied = result.Proxied
	record.TTL = result.TTL
	return record
}

// NewDate returns today as string
func NewDate() string {
	t := time.Now()
	return t.Format("2006-01-02")
}

func RandomNumber() string {
	// seed time
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	// generate random number
	return fmt.Sprintf("%d", r.Intn(999))
}

// RemoveRestrictedSubdomains removes restricted subdomains from the list in restricted.json
func RemoveRestrictedSubdomains(filename string, localRecords []schema.Record) (localNonRestrictedRecords []schema.Record, localRestrictedRecords []schema.Record) {
	restrictedRecords := ReadRestrictedRecords(filename)
	for _, record := range localRecords {
		if !IsRestricted(record.Name, restrictedRecords) {
			localNonRestrictedRecords = append(localNonRestrictedRecords, record)
		} else {
			localRestrictedRecords = append(localRestrictedRecords, record)
		}
	}
	return localNonRestrictedRecords, localRestrictedRecords
}

// IsRestricted checks if the record is restricted
func IsRestricted(name string, restrictedRecords []string) bool {
	for _, record := range restrictedRecords {
		// check using regular expression
		if regexp.MustCompile(record).MatchString(name) {
			return true
		}
	}
	return false
}

// ReadRestrictedRecords read restricted records from restricted.json and store in a array
func ReadRestrictedRecords(filename string) []string {
	type Restricted struct {
		RestrictedSubdomain []string `json:"restricted_subdomain"`
	}
	restrictedRecords := Restricted{}
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(file, &restrictedRecords)
	if err != nil {
		fmt.Println(err)
	}
	return restrictedRecords.RestrictedSubdomain
}

// ConfirmPrompt will prompt to user for yes or no
func ConfirmPrompt(message string) bool {
	var response string
	fmt.Print(message + " (yes/no) :")
	fmt.Scanln(&response)

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return false
	}
}
