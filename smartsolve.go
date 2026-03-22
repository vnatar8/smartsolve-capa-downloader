package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CAPA holds a CAPA record from SmartSolve.
type CAPA struct {
	Number string
	Title  string
	Site   string
	Status string
	Phase  string
	Owner  string
}

// searchRequest is the POST body for the getsearchdata API.
type searchRequest struct {
	EntityName             string                 `json:"entityName"`
	CurrentPageIndex       int                    `json:"currentPageIndex"`
	PageSize               int                    `json:"pageSize"`
	Filter                 []interface{}          `json:"filter"`
	DynamicParamCollection map[string]interface{} `json:"dynamicParamCollection"`
	SavedSearchName        string                 `json:"savedSearchName"`
}

// searchResponse is the JSON response from getsearchdata.
type searchResponse struct {
	SearchData struct {
		Rows []struct {
			RecordNumber string `json:"RECORD_NUMBER"`
			Title        string `json:"TITLE"`
			Site         string `json:"OWNING_SITE_CODE"`
			Status       string `json:"STATUS_UNTRANSLATED"`
			Phase        string `json:"CURRENT_PHASE"`
			Owner        string `json:"CAPA_OWNER_CODE"`
		} `json:"rows"`
		TotalResultCount string `json:"totalResultCount"`
	} `json:"SearchData"`
	HasError bool `json:"hasError"`
}

// newSmartSolveRequest creates an authenticated HTTP request.
func newSmartSolveRequest(method, url string, body io.Reader, token string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", config.BaseURL)
	req.Header.Set("Referer", config.BaseURL+"/")
	return req, nil
}

// fetchAllCAPAs queries SmartSolve for all CAPAs.
func fetchAllCAPAs(token string) ([]CAPA, error) {
	reqBody := searchRequest{
		EntityName:             "CAPA",
		CurrentPageIndex:       0,
		PageSize:               500,
		Filter:                 []interface{}{},
		DynamicParamCollection: map[string]interface{}{},
		SavedSearchName:        "All CAPAs",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	endpoint := config.APIURL + "apis/getsearchdata"
	req, err := newSmartSolveRequest("POST", endpoint, bytes.NewReader(bodyBytes), token)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed (HTTP %d); please log into SmartSolve in Chrome and try again", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned HTTP %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	var result searchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("cannot parse response JSON: %w", err)
	}

	if result.HasError {
		return nil, fmt.Errorf("SmartSolve API returned an error")
	}

	var capas []CAPA
	for _, row := range result.SearchData.Rows {
		capas = append(capas, CAPA{
			Number: row.RecordNumber,
			Title:  row.Title,
			Site:   row.Site,
			Status: row.Status,
			Phase:  row.Phase,
			Owner:  row.Owner,
		})
	}

	return capas, nil
}

// filterCAPAs filters CAPAs by site and status mode.
func filterCAPAs(capas []CAPA, site string, mode string) (matched []CAPA, skipped []CAPA) {
	for _, c := range capas {
		if site != "" && c.Site != site {
			continue
		}

		switch mode {
		case "open":
			if c.Status == "CLOSED" || c.Status == "VOID" {
				skipped = append(skipped, c)
			} else {
				matched = append(matched, c)
			}
		case "closed":
			if c.Status == "CLOSED" {
				matched = append(matched, c)
			} else {
				skipped = append(skipped, c)
			}
		case "all":
			matched = append(matched, c)
		}
	}
	return
}

// testSession verifies the JWT token is valid by making a small API call.
func testSession(token string) error {
	reqBody := searchRequest{
		EntityName:             "CAPA",
		CurrentPageIndex:       0,
		PageSize:               1,
		Filter:                 []interface{}{},
		DynamicParamCollection: map[string]interface{}{},
		SavedSearchName:        "All CAPAs",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	endpoint := config.APIURL + "apis/getsearchdata"
	req, err := newSmartSolveRequest("POST", endpoint, bytes.NewReader(bodyBytes), token)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to SmartSolve: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("session expired (HTTP %d); please log into SmartSolve in Chrome and try again", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response: HTTP %d", resp.StatusCode)
	}
	return nil
}
