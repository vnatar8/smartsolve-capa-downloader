package main

// SmartSolve instance configuration.
// Update these values to match your organization's SmartSolve deployment.
var config = struct {
	// BaseURL is your SmartSolve portal URL (used for Origin/Referer headers).
	// Example: "https://mycompany.pilgrimasp.com"
	BaseURL string

	// APIURL is the SmartSolve WOPI API base URL (used for API calls and PDF downloads).
	// Example: "https://mycompany.wopi.pilgrimasp.com/prod/smartsolve/"
	APIURL string

	// SiteCode is your site's code in SmartSolve (used to filter CAPAs).
	// Set to "" to download CAPAs from all sites.
	// Example: "NYC", "LON", "MEL"
	SiteCode string

	// ReportName is the SmartSolve report type for CAPA detail PDFs.
	// This is typically "CAPA_DETAIL" and should not need changing.
	ReportName string
}{
	BaseURL:    "https://CHANGEME.pilgrimasp.com",
	APIURL:     "https://CHANGEME.wopi.pilgrimasp.com/prod/smartsolve/",
	SiteCode:   "CHANGEME",
	ReportName: "CAPA_DETAIL",
}
