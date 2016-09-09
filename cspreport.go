package cspreport

// CSPReport is a Content Security Policy report as per http://www.w3.org/TR/CSP/
type CSPReport struct {
	Body struct {
		DocumentURI        string `json:"document-uri"`
		Referrer           string `json:"referrer"`
		BlockedURI         string `json:"blocked-uri"`
		ViolatedDirective  string `json:"violated-directive"`
		EffectiveDirective string `json:"effective-directive"`
		OriginalPolicy     string `json:"original-policy"`
		Timestamp          string `json:"@timestamp"`
	} `json:"csp-report"`
}

// PKPReport is a Public Key Pinning report as per https://tools.ietf.org/html/draft-ietf-websec-key-pinning-21
type PKPReport struct {
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

// BatchReportStorage is a way to store incoming reports
type BatchReportStorage interface {
	Start() error
	Stop() error
	AddCSPReport(CSPReport)
	AddPKPReport(PKPReport)
}

// Logger is a simple logging wrapper interface
type Logger interface {
	Log(...interface{}) error
}
