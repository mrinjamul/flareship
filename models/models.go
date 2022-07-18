package models

// Record is the struct for the record
type Record struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Content   string `json:"content,omitempty"`
	Proxiable bool   `json:"proxiable,omitempty"`
	Proxied   bool   `json:"proxied,omitempty"`
	TTL       uint   `json:"ttl,omitempty"`
}

// Owner is the struct for the owner schema
type Owner struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

// Records is the struct for the records which is parsed from file
type Records struct {
	Description string `json:"description,omitempty"`
	Repo        string `json:"repo,omitempty"`
	Owner       Owner  `json:"owner,omitempty"`
	Record      Record `json:"record"`
}

// Result is the record which is returned from the API
type Result struct {
	ID         string `json:"id"`
	ZoneID     string `json:"zone_id"`
	ZoneName   string `json:"zone_name"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Proxiable  bool   `json:"proxiable"`
	Proxied    bool   `json:"proxied"`
	TTL        uint   `json:"ttl"`
	CreatedOn  string `json:"created_on"`
	ModifiedOn string `json:"modified_on"`
}

// ResultInfo is the status of the request
type ResultInfo struct {
	Page       uint `json:"page"`
	PerPage    uint `json:"per_page"`
	Count      uint `json:"count"`
	TotalCount uint `json:"total_count"`
	TotalPages uint `json:"total_pages"`
}

// ErrorChain is the error chain
type ErrorChain struct {
	Code uint   `json:"code"`
	Type string `json:"type"`
}

// Error is the error struct
type Errors struct {
	Code       uint       `json:"code"`
	Message    string     `json:"message"`
	ErrorChain ErrorChain `json:"error_chain"`
}

// CFResponse is the response struct we get from the API using GET method
type CFResponse struct {
	Success    bool       `json:"success"`
	Errors     []Errors   `json:"errors"`
	Messages   []string   `json:"messages"`
	ResultInfo ResultInfo `json:"result_info"`
	Result     []Result   `json:"result"`
}

// CFResponse is the response struct we get from the API using POST method
type PostResponse struct {
	Success    bool         `json:"success"`
	Errors     []Errors     `json:"errors"`
	Messages   []string     `json:"messages"`
	ResultInfo []ResultInfo `json:"result_info"`
	Result     Result       `json:"result"`
}

// DelResult is the result struct for the DELETE request
type DelResult struct {
	ID string `json:"id"`
}

// DelResponse is the response struct we get from the API using DELETE method
type DelResponse struct {
	Result DelResult `json:"result"`
	Errors []Errors  `json:"errors"`
}

// Config is the struct for the config file
type Config struct {
	CFToken        string   `json:"cf_token"`
	ZoneID         string   `json:"zone_id"`
	DomainName     string   `json:"domain_name"`
	RecordFile     string   `json:"record_file"`
	RestrictedFile string   `json:"restricted_file"`
	RecordType     []string `json:"record_type"`
}
