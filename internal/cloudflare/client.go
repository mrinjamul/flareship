package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/mrinjamul/flareship/internal/utils"
	"github.com/mrinjamul/flareship/pkg/schema"
)

var (
	// BaseAPI is the base url for cloudflare api
	BaseAPI string = "https://api.cloudflare.com/client/v4/"
)

// ReadRecord creates a GET request
func ReadRecord(zoneID, query, token string) (schema.CFResponse, error) {
	var result schema.CFResponse
	endpoint := "zones/" + zoneID + "/dns_records?" + query
	url := BaseAPI + endpoint
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(body)
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return schema.CFResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}

// GetRecords returns all records from cloudflare api
func ReadAllRecords(zoneID, token string, recordTypes []string) []schema.Record {
	query := url.Values{}
	var records []schema.Record
	var results []schema.Result
	for _, t := range recordTypes {
		query.Add("type", t)
		perPage := 100
		page := 1
		query.Add("per_page", strconv.Itoa(perPage))
		for ok := true; ok; ok = (len(results) == perPage) {
			query.Add("page", strconv.Itoa(page))
			query := query.Encode()
			resp, err := ReadRecord(zoneID, query, token)
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
func CreateRecord(zoneID, token string, postBody []byte) schema.Record {
	var result schema.Result
	endpoint := "zones/" + zoneID + "/dns_records"
	resp, err := httpPost("POST", endpoint, postBody, token)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to create records")
		os.Exit(1)
	}
	result = resp.Result
	record := utils.ConcatOne(schema.Record{}, result)
	return record
}

// UpdateRecord updates a record
func UpdateRecord(zoneID, token, recordID string, postBody []byte) schema.Record {
	var result schema.Result
	endpoint := "zones/" + zoneID + "/dns_records/" + recordID
	resp, err := httpPost("PUT", endpoint, postBody, token)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to update records")
		os.Exit(1)
	}
	result = resp.Result
	record := utils.ConcatOne(schema.Record{}, result)
	return record
}

// DeleteRecord delete a record
func DeleteRecord(zoneID, token, recordID string) schema.DelResponse {
	endpoint := "zones/" + zoneID + "/dns_records/" + recordID
	resp, err := httpDelete(endpoint, token)
	if err != nil {
		fmt.Println(err)
		fmt.Println("ERROR - fail to delete records")
		os.Exit(1)
	}
	return resp
}

// httpPost creates a POST or PUT or PATCH request
func httpPost(method string, endpoint string, postBody []byte, token string) (schema.PostResponse, error) {
	var result schema.PostResponse
	if method == "" {
		method = "POST"
	}
	url := BaseAPI + endpoint
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return schema.PostResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}

// httpDelete creates a DELETE request
func httpDelete(endpoint string, token string) (schema.DelResponse, error) {
	var result schema.DelResponse
	url := BaseAPI + endpoint
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response bytes:", err)
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error while parsing the response bytes:", err)
	}
	if len(result.Errors) > 0 {
		return schema.DelResponse{}, fmt.Errorf("%s", result.Errors[0].Message)
	}
	return result, nil
}
