package frontreport

import "strings"

// Reportable structs can be saved to Elastic
// - their Type and Service defines index to save to;
// - they can hold Timestamp for Elastic to sort on.
type Reportable interface {
	GetType() string
	GetService() string
	SetTimestamp(string)
	SetHost(string)
}

// Report is a generic report type (they don't have much in common)
type Report struct {
	Timestamp string `json:"@timestamp"`
	Host      string `json:"frontreport-host"`
	Service   string `json:"service"`
}

// SetTimestamp sets timestamp for Elastic default sorting
func (r *Report) SetTimestamp(ts string) {
	r.Timestamp = ts
}

// SetHost sets hostname to tell apart reports from different sites
func (r *Report) SetHost(h string) {
	r.Host = h
}

// GetService returns service to tell apart reports from different sites
func (r *Report) GetService() string {
	return strings.ToLower(r.Service)
}

// CSPReport is a Content Security Policy report as per http://www.w3.org/TR/CSP/
type CSPReport struct {
	Report
	Body struct {
		DocumentURI        string `json:"document-uri"`
		Referrer           string `json:"referrer"`
		BlockedURI         string `json:"blocked-uri"`
		ViolatedDirective  string `json:"violated-directive"`
		EffectiveDirective string `json:"effective-directive"`
		OriginalPolicy     string `json:"original-policy"`
	} `json:"csp-report"`
}

// GetType returns report type
func (c CSPReport) GetType() string {
	return "csp"
}

// PKPReport is a Public Key Pinning report as per https://tools.ietf.org/html/draft-ietf-websec-key-pinning-21
type PKPReport struct {
	Report
	DateTime                  string   `json:"date-time"`
	Hostname                  string   `json:"hostname"`
	Port                      int      `json:"port"`
	EffectiveExpirationDate   string   `json:"effective-expiration-date"`
	IncludeSubdomains         bool     `json:"include-subdomains"`
	NotedHostname             string   `json:"noted-hostname"`
	ServedCertificateChain    []string `json:"served-certificate-chain"`
	ValidatedCertificateChain []string `json:"validated-certificate-chain"`
	KnownPins                 []string `json:"known-pins"`
}

// GetType returns report type
func (p PKPReport) GetType() string {
	return "pkp"
}

// StacktraceJSStackframe is a single stack frame representation
type StacktraceJSStackframe struct {
	FunctionName string `json:"functionName"`
	FileName     string `json:"fileName"`
	LineNumber   int    `json:"lineNumber"`
	ColumnNumber int    `json:"columnNumber"`
}

// StacktraceJSReport is a universal browser stacktrace format as per https://github.com/stacktracejs/stacktrace.js#stacktracereportstackframes-url-message--promisestring
type StacktraceJSReport struct {
	Report
	Message string                   `json:"message"`
	Stack   []StacktraceJSStackframe `json:"stack"`

	// These fields are not a part of StacktraceJS specification, but are useful for error reports
	Browser *struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
	} `json:"browser,omitempty"`
	OS *struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
	} `json:"os,omitempty"`
	URL       string `json:"url,omitempty"`
	SourceURL string `json:"sourceUrl,omitempty"`
	UserID    string `json:"userId,omitempty"`
	StackHash string `json:"stackHash,omitempty"`
	ScriptURL string `json:"scriptUrl,omitempty"`

	// These fields are experimental and can disappear in the future
	PartyID         string `json:"partyId,omitempty"`
	DepartmentID    string `json:"departmentId,omitempty"`
	SalesPointID    string `json:"salesPointId,omitempty"`
	RetailUIVersion string `json:"retailUiVersion,omitempty"`
	AppVersion      string `json:"appVersion,omitempty"`
	ClaimID         string `json:"claimId,omitempty"`
}

// GetType returns report type
func (s StacktraceJSReport) GetType() string {
	return "stacktracejs"
}

// ReportStorage is a way to store incoming reports
type ReportStorage interface {
	AddReport(Reportable)
}

// SourcemapProcessor converts stacktrace to readable format using sourcemaps
type SourcemapProcessor interface {
	ProcessStack([]StacktraceJSStackframe) []StacktraceJSStackframe
}
