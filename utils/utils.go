package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mrinjamul/flareship/models"
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
		// "Use `flareship config --gen` to generate config file",
		"Use `flareship sync --dry-run` to see what will be synced",
		"Use `flareship sync` to sync your records",
		"Use `flareship sync --domain [url]` to specify the root domain",
		"Use `flareship sync -f [record_file]` to specify the file to sync",
		"Use `flareship fmt --check` to check records file",
		"Use `flareship fmt` to format records file",
		"Use `flareship fmt --domain [url]` to specify the root domain",
		// "Use `flareship fmt --dry-run` to see what will be formatted",
	}
	rand.Seed(time.Now().UnixNano())
	return tips[rand.Intn(len(tips))]
}

// GenerateConfig generates the config file
func GenerateConfig(filename string) error {
	recordFile := HomeDir() + "/.mrinjamulcli_dns_records.json"
	if _, err := os.Stat("records.json"); err == nil {
		// get current path
		path, _ := os.Getwd()
		recordFile = path + "/" + "records.json"
	}
	config := models.Config{
		DomainName: "",
		RecordFile: recordFile,
		CFToken:    "",
		ZoneID:     "",
		RecordType: []string{"A", "CNAME"},
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	if _, err := os.Stat(recordFile); os.IsNotExist(err) {
		records := []models.Records{
			{
				Description: "This is a sample record",
				Repo:        "",
				Owner: models.Owner{
					Username: "username",
					Email:    "username@domain.com",
				},
				Record: models.Record{
					Type:      "A",
					Name:      "*.dev",
					Content:   "127.0.0.1",
					Proxiable: true,
					Proxied:   true,
					TTL:       1,
				},
			},
		}
		data, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(recordFile, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseConfig parses the config file
func ParseConfig(filename string) (models.Config, error) {
	var config models.Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// GetConfig returns the config variables
func GetConfig(filename string) (DomainName, RecordFile, RestrictedFile, CFToken, ZoneID string, RecordType []string) {
	// check if config file exists
	if filename == "" {
		filename = HomeDir() + "/.mrinjamulcli.json"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// fmt.Println("Config file not found. Please run `flareship config --gen` to generate config file")
			// GenerateConfig(filename)
			return "", "", "", "", "", []string{}
		}
	}
	config, err := ParseConfig(filename)
	if err != nil {
		fmt.Println("ERROR - fail to parse config file")
		GenerateConfig(filename)
		os.Exit(1)
	}
	return config.DomainName, config.RecordFile, config.RestrictedFile, config.CFToken, config.ZoneID, config.RecordType
}

// GetRecords parse records from records file
func GetRecords(filename string) ([]models.Records, error) {
	var records []models.Records
	data, err := os.ReadFile(filename)
	if err != nil {
		return []models.Records{}, err
	}
	err = json.Unmarshal(data, &records)
	if err != nil {

		return []models.Records{}, err
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
func GetDNSRecords(filename string, enabledRecordType []string) ([]models.Record, error) {
	var records []models.Record
	entries, err := GetRecords(filename)
	if err != nil {
		return []models.Record{}, err
	}
	for _, entry := range entries {
		if TypeContains(enabledRecordType, entry.Record.Type) {
			records = append(records, entry.Record)
		}
	}
	return records, nil
}

// FindRecordByName returns the record from name
func FindRecordByName(records []models.Record, name string) models.Record {
	for _, record := range records {
		if record.Name == name {
			return record
		}
	}
	return models.Record{}
}

// FindRecordID returns the record ID from name
func FindRecordID(records []models.Record, name string) string {
	for _, r := range records {
		if r.Name == name {
			return r.ID
		}
	}
	return ""
}

// RecordContain checks if a single record is in the records
func RecordContain(records []models.Record, record models.Record) bool {
	for _, r := range records {
		if r.Name == record.Name {
			return true
		}
	}
	return false
}

// RecordContains checks if the sub-record is in the records
func RecordContains(records []models.Record, subrecords []models.Record) bool {
	for _, r := range subrecords {
		if !RecordContain(records, r) {
			return false
		}
	}
	return true
}

// CFFetch creates a GET request
func CFFetch(base string, endpoint string, token string) (models.CFResponse, error) {
	var result models.CFResponse
	url := base + endpoint
	// Create a Bearer string by appending string access token
	bearer := "Bearer " + token
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on response.\nERROR -", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(body)
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return models.CFResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}

// CFPost creates a POST or PUT or PATCH request
func CFPost(method string, base string, endpoint string, postBody []byte, token string) (models.PostResponse, error) {
	var result models.PostResponse
	if method == "" {
		method = "POST"
	}
	url := base + endpoint
	responseBody := bytes.NewBuffer(postBody)
	// Create a Bearer string by appending string access token
	bearer := "Bearer " + token
	// Create a new request using http
	req, err := http.NewRequest(method, url, responseBody)
	if err != nil {
		fmt.Println(err)
	}
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on response.\nERROR -", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return models.PostResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}

// CFDelete creates a DELETE request
func CFDelete(base string, endpoint string, token string) (models.DelResponse, error) {
	var result models.DelResponse
	url := base + endpoint
	// Create a Bearer string by appending string access token
	bearer := "Bearer " + token
	// Create a new request using http
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on response.\nERROR -", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return models.DelResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}

// Concat converts results to records
func Concat(records []models.Record, result []models.Result) []models.Record {
	for _, r := range result {
		// record := make(models.Record, 0)
		var record models.Record
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
func ConcatOne(record models.Record, result models.Result) models.Record {
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
	rand.Seed(time.Now().UnixNano())
	// generate random number
	return fmt.Sprintf("%d", rand.Intn(999))
}

// RemoveRestrictedSubdomains removes restricted subdomains from the list in restricted.json
func RemoveRestrictedSubdomains(filename string, localRecords []models.Record) (localNonRestrictedRecords []models.Record, localRestrictedRecords []models.Record) {
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
	file, err := ioutil.ReadFile(filename)
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
