# SmartSolve CAPA Downloader

A CLI tool that batch-downloads CAPA Detail PDFs from [IQVIA SmartSolve](https://www.iqvia.com/solutions/compliance/smartsolve) eQMS. It reads your authentication token directly from Chrome's local storage, so you don't need to interact with the browser at all. Just be logged into SmartSolve in Chrome and run the tool.

## Features

- Automatically reads your JWT auth token from Chrome (no manual copy-paste needed in most cases)
- Downloads CAPA Detail PDFs concurrently (5 at a time)
- Filters by site and status (open, closed, or all)
- Validates downloaded PDFs (checks file headers and size)
- Handles duplicate filenames with automatic suffixes
- Retries failed downloads once before reporting them
- Single binary with zero runtime dependencies

## Prerequisites

- **Windows** (uses Windows APIs for reading Chrome's locked database files)
- **Google Chrome** with an active SmartSolve session (you must be logged in)
- **Go 1.21+** (only needed if building from source)

## Setup

### 1. Clone and configure

```
git clone https://github.com/<your-username>/smartsolve-capa-downloader.git
cd smartsolve-capa-downloader
```

Edit `config.go` with your organization's SmartSolve details:

```go
var config = struct {
    BaseURL    string
    APIURL     string
    SiteCode   string
    ReportName string
}{
    BaseURL:    "https://yourcompany.pilgrimasp.com",
    APIURL:     "https://yourcompany.wopi.pilgrimasp.com/prod/smartsolve/",
    SiteCode:   "NYC",       // your site code
    ReportName: "CAPA_DETAIL",
}
```

**How to find your URLs:**
1. Log into SmartSolve in Chrome
2. Your browser address bar shows the BaseURL (e.g., `https://yourcompany.pilgrimasp.com/prod/...`)
3. Open DevTools (F12), go to Network tab, navigate to a CAPA list, and look at the API request URLs for the APIURL

**How to find your site code:**
Look at the "Site" column in the SmartSolve CAPA list. Common codes: three-letter abbreviations like `NYC`, `LON`, `MEL`, `RCH`.

### 2. Build

```
go build -o capa-downloader.exe .
```

### 3. Distribute

Share the compiled `capa-downloader.exe` with your team. No Go installation or dependencies needed to run it.

## Usage

```
capa-downloader.exe --output <directory> [flags]
```

### Required

- `--output <path>` : Directory to save downloaded PDFs. Must already exist.

### Optional

- `--closed` : Download only Closed CAPAs (default: non-Closed, non-Void)
- `--all` : Download all CAPAs regardless of status
- `--capa <number>` : Download a single specific CAPA (e.g., `CAPA-2025-000043`)
- `--site <code>` : Override the site code from config (e.g., `--site LON`)
- `--token <jwt>` : Provide a JWT token manually (bypasses Chrome auto-detection)

### Examples

Download all open CAPAs for your site:
```
capa-downloader.exe --output "C:\Downloads\CAPAs"
```

Download all closed CAPAs:
```
capa-downloader.exe --output "C:\Downloads\CAPAs" --closed
```

Download one specific CAPA:
```
capa-downloader.exe --output "C:\Downloads\CAPAs" --capa CAPA-2025-000043
```

Download everything (open, closed, void):
```
capa-downloader.exe --output "C:\Downloads\CAPAs" --all
```

Download CAPAs from a different site:
```
capa-downloader.exe --output "C:\Downloads\CAPAs" --site LON
```

## How It Works

1. **Reads your JWT token** from Chrome's Local Storage LevelDB files on disk. SmartSolve stores the auth token in `localStorage` under the key `token`. The tool scans the LevelDB data files for this value without needing to open Chrome or interact with the browser.

2. **Queries SmartSolve's API** (`POST /apis/getsearchdata`) to get the full CAPA list, then filters by your site code and the requested status.

3. **Downloads PDFs concurrently** using the same JWT token. Up to 5 downloads run in parallel.

4. **Validates each PDF** by checking the `%PDF-` magic bytes and minimum file size. Invalid responses (error pages, expired sessions) are caught and reported.

5. **Saves to the output folder** with filenames like `CAPA-2025-000043.pdf`. If a file already exists, it appends `_1`, `_2`, etc.

## If Auto-Detection Fails

If the tool can't find your token automatically (this can happen if Chrome hasn't flushed localStorage to disk recently), it will print instructions:

```
Could not find SmartSolve token automatically.

To get your token, open SmartSolve in Chrome, press F12, go to Console, and run:

  copy(localStorage.getItem("token"))

Then run this tool with --token <paste>
```

The token is valid for 24 hours. You only need to do this once per day.

## Project Structure

```
├── config.go              # SmartSolve instance configuration (URLs, site code)
├── main.go                # CLI entry point, flag parsing, orchestration
├── sessionstorage.go      # Reads JWT token from Chrome's Local Storage
├── smartsolve.go          # SmartSolve API client (CAPA list query, filtering)
├── downloader.go          # Concurrent PDF download with validation and retry
├── fileutil.go            # File saving with duplicate filename handling
├── *_test.go              # Tests
├── go.mod / go.sum        # Go module dependencies
└── README.md
```

## Running Tests

```
go test ./... -v
```

Note: Some tests require Chrome to be installed and SmartSolve to be logged in. These will be skipped automatically if the prerequisites aren't met.

## Security Notes

- This tool is **read-only**. It never creates, modifies, or deletes any record in SmartSolve.
- JWT tokens are read from Chrome's local files and held in memory only for the duration of the download session. They are never written to disk or transmitted anywhere other than to your SmartSolve instance.
- The tool accesses Chrome's Local Storage database files directly. It does not intercept network traffic or inject into Chrome processes.
- No credentials (usernames, passwords) are ever handled by this tool. Authentication is managed entirely by your browser session.

## Limitations

- **Windows only.** The tool uses Windows-specific APIs (`CreateFile` with shared access flags) to read Chrome's locked database files. macOS/Linux support would require different file-access strategies.
- **Chrome only.** Other browsers (Edge, Firefox) store localStorage differently. Edge (Chromium-based) may work with a modified storage path, but this is untested.
- **SmartSolve-specific JWT format.** The auto-detection looks for JWTs with the header `{"typ":"JWT","alg":"HS256"}`. If your SmartSolve instance uses a different signing algorithm, you may need to update the `targetPrefix` in `sessionstorage.go` or use `--token` manually.

## License

MIT
