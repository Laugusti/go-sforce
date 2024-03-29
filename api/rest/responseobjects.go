package restapi

// UpsertResult is a successful response from the Salesforce API after an upsert.
type UpsertResult struct {
	ID      string        `json:"id"`
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
}

// QueryResult is successful response from the Salesforce API after a query.
type QueryResult struct {
	TotalSize      int       `json:"totalSize"`
	Done           bool      `json:"done"`
	NextRecordsURL string    `json:"nextRecordsURL,omitempty"`
	Records        []SObject `json:"records"`
}
