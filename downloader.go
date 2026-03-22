package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

const maxConcurrent = 5

// DownloadResult holds the outcome of a single CAPA download.
type DownloadResult struct {
	CAPANumber string
	Data       []byte
	Err        error
}

// validatePDF checks that data looks like a valid PDF.
func validatePDF(data []byte) error {
	if len(data) < 1024 {
		return fmt.Errorf("file too small (%d bytes); likely an error page", len(data))
	}
	if len(data) < 5 || string(data[:5]) != "%PDF-" {
		return fmt.Errorf("not a PDF (starts with %q)", string(data[:min(20, len(data))]))
	}
	return nil
}

// downloadCAPADetail downloads a single CAPA Detail PDF with one retry on failure.
func downloadCAPADetail(client *http.Client, token string, capaNumber string) ([]byte, error) {
	reqURL := fmt.Sprintf(
		"%sV2SmartReport.aspx?RptName=%s&RECORD_NUMBER=%s&RECORD_REVISION=&Token=%s",
		config.APIURL, config.ReportName, url.QueryEscape(capaNumber), url.QueryEscape(token),
	)

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			fmt.Printf("  Retrying %s ...\n", capaNumber)
		}

		resp, err := client.Get(reqURL)
		if err != nil {
			lastErr = fmt.Errorf("download request failed: %w", err)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		if err := validatePDF(data); err != nil {
			lastErr = err
			continue
		}

		return data, nil
	}
	return nil, lastErr
}

// downloadAll downloads multiple CAPA Detail PDFs concurrently.
func downloadAll(client *http.Client, token string, capas []string) []DownloadResult {
	results := make([]DownloadResult, len(capas))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, capa := range capas {
		wg.Add(1)
		go func(idx int, capaNum string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("  Downloading %s ...\n", capaNum)
			data, err := downloadCAPADetail(client, token, capaNum)
			results[idx] = DownloadResult{
				CAPANumber: capaNum,
				Data:       data,
				Err:        err,
			}
			if err != nil {
				fmt.Printf("  FAILED %s: %v\n", capaNum, err)
			} else {
				fmt.Printf("  OK %s (%d bytes)\n", capaNum, len(data))
			}
		}(i, capa)
	}

	wg.Wait()
	return results
}
