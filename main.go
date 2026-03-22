package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outputDir := flag.String("output", "", "Directory to save downloaded PDFs (required)")
	closed := flag.Bool("closed", false, "Download only Closed CAPAs")
	all := flag.Bool("all", false, "Download all CAPAs regardless of status")
	capaNum := flag.String("capa", "", "Download a single specific CAPA by number (any status)")
	token := flag.String("token", "", "JWT token (optional; bypasses reading from Chrome)")
	site := flag.String("site", config.SiteCode, "Site code to filter CAPAs (e.g. NYC, LON)")

	flag.Parse()

	// Validate configuration
	if strings.Contains(config.BaseURL, "CHANGEME") {
		fmt.Fprintln(os.Stderr, "Error: you must configure your SmartSolve URLs in config.go before building.")
		fmt.Fprintln(os.Stderr, "See README.md for setup instructions.")
		os.Exit(1)
	}

	// Validate output directory
	if *outputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --output is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, `  capa-downloader --output "C:\Downloads\CAPAs"`)
		fmt.Fprintln(os.Stderr, `  capa-downloader --output "C:\Downloads\CAPAs" --closed`)
		fmt.Fprintln(os.Stderr, `  capa-downloader --output "C:\Downloads\CAPAs" --capa CAPA-2025-000043`)
		os.Exit(1)
	}

	absPath, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid output path: %v\n", err)
		os.Exit(1)
	}
	*outputDir = absPath

	if info, err := os.Stat(*outputDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: output directory does not exist: %s\n", *outputDir)
		os.Exit(1)
	}

	// Determine filter mode
	mode := "open"
	if *closed {
		mode = "closed"
	}
	if *all {
		mode = "all"
	}

	// Step 1: Get JWT token
	jwtToken := *token
	if jwtToken == "" {
		fmt.Println("Reading JWT token from Chrome session...")
		var err error
		jwtToken, err = readJWTFromLocalStorage()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not find SmartSolve token automatically.")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "To get your token, open SmartSolve in Chrome, press F12, go to Console, and run:")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, `  copy(localStorage.getItem("token"))`)
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Then run this tool with --token <paste>")
			os.Exit(1)
		}
		fmt.Printf("Found token (%d chars).\n", len(jwtToken))
	} else {
		fmt.Println("Using provided JWT token.")
	}

	// Step 2: Verify session
	fmt.Println("Verifying SmartSolve session...")
	if err := testSession(jwtToken); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Session valid.")

	// Step 3: Get CAPA list
	var capaNumbers []string
	var skippedCount int

	if *capaNum != "" {
		capaNumbers = []string{*capaNum}
		fmt.Printf("Single CAPA mode: %s\n", *capaNum)
	} else {
		fmt.Printf("Fetching CAPA list from SmartSolve (mode: %s, site: %s)...\n", mode, *site)
		allCAPAs, err := fetchAllCAPAs(jwtToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d total CAPAs.\n", len(allCAPAs))

		matched, skipped := filterCAPAs(allCAPAs, *site, mode)
		skippedCount = len(skipped)

		if len(matched) == 0 {
			fmt.Printf("No CAPAs found matching criteria (site=%s, mode=%s).\n", *site, mode)
			os.Exit(0)
		}

		for _, c := range matched {
			capaNumbers = append(capaNumbers, c.Number)
		}
		fmt.Printf("Downloading %d CAPAs, skipped %d.\n", len(matched), skippedCount)
	}

	// Step 4: Download PDFs
	fmt.Printf("\nDownloading %d CAPA Detail PDFs...\n", len(capaNumbers))
	client := &http.Client{}
	results := downloadAll(client, jwtToken, capaNumbers)

	// Step 5: Save results
	var downloaded []string
	var failed []string
	for _, r := range results {
		if r.Err != nil {
			failed = append(failed, fmt.Sprintf("  %s - %v", r.CAPANumber, r.Err))
			continue
		}
		path, err := savePDF(*outputDir, r.CAPANumber, r.Data)
		if err != nil {
			failed = append(failed, fmt.Sprintf("  %s - %v", r.CAPANumber, err))
			continue
		}
		downloaded = append(downloaded, fmt.Sprintf("  %s", filepath.Base(path)))
	}

	// Step 6: Print summary
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("CAPA Detail Download Complete")
	fmt.Println("========================================")
	fmt.Printf("\nDownloaded (%d):\n", len(downloaded))
	for _, d := range downloaded {
		fmt.Println(d)
	}
	if skippedCount > 0 {
		fmt.Printf("\nSkipped (%d, filtered out by status).\n", skippedCount)
	}
	if len(failed) > 0 {
		fmt.Printf("\nFailed (%d):\n", len(failed))
		for _, f := range failed {
			fmt.Println(f)
		}
	}
	fmt.Printf("\nAll files saved to: %s\n", *outputDir)
}
