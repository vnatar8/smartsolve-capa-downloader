package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// chromeLocalStoragePath returns the path to Chrome's Local Storage LevelDB directory.
func chromeLocalStoragePath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return ""
	}
	return filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Local Storage", "leveldb")
}

// readJWTFromLocalStorage reads the SmartSolve JWT token from Chrome's Local Storage.
// Chrome stores localStorage in a LevelDB database on disk. Rather than using a LevelDB
// library (which can crash on Windows with Chrome's locked files), this function scans
// the raw .ldb and .log files for the JWT token string directly.
func readJWTFromLocalStorage() (string, error) {
	lsPath := chromeLocalStoragePath()
	if lsPath == "" {
		return "", fmt.Errorf("LOCALAPPDATA not set")
	}
	if _, err := os.Stat(lsPath); err != nil {
		return "", fmt.Errorf("Chrome Local Storage not found at %s", lsPath)
	}

	// SmartSolve JWT header: {"typ":"JWT","alg":"HS256"} = eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9
	// Chrome stores localStorage values as UTF-16, so we search for the
	// UTF-16 encoded version of this header prefix.
	targetPrefix := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"

	var token string
	err := filepath.Walk(lsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ldb" && ext != ".log" {
			return nil
		}

		data, readErr := readFileShared(path)
		if readErr != nil {
			return nil
		}

		// Search in plain ASCII first
		content := string(data)
		if idx := strings.Index(content, targetPrefix); idx >= 0 {
			extracted := extractJWT(content[idx:])
			if len(extracted) > 100 {
				token = extracted
				return filepath.SkipAll
			}
		}

		// Search in UTF-16 (every ASCII char followed by a \0 byte)
		found := findUTF16JWT(data, targetPrefix)
		if found != "" && len(found) > 100 {
			token = found
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("error scanning Local Storage: %w", err)
	}

	if token == "" {
		return "", fmt.Errorf("no SmartSolve JWT token found in Chrome Local Storage; make sure you are logged into SmartSolve in Chrome")
	}

	return token, nil
}

// readFileShared reads a file using Windows shared access flags,
// allowing reading even when Chrome has the file open.
func readFileShared(path string) ([]byte, error) {
	pathUTF16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	handle, err := windows.CreateFile(
		pathUTF16,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, err
	}

	// Wrap the handle in an os.File and read from it directly.
	// Do NOT call os.ReadFile(path) as that would open a new handle without shared access.
	f := os.NewFile(uintptr(handle), path)
	defer f.Close()

	return io.ReadAll(f)
}

// extractJWT pulls the JWT string out of a value that may contain surrounding
// characters. JWTs start with "eyJ" and consist only of Base64URL characters and dots.
func extractJWT(value string) string {
	idx := strings.Index(value, "eyJ")
	if idx < 0 {
		return ""
	}
	token := value[idx:]
	end := len(token)
	for i, ch := range token {
		if !isJWTChar(byte(ch)) {
			end = i
			break
		}
	}
	return token[:end]
}

// findUTF16JWT searches binary data for a UTF-16LE encoded JWT token
// matching the given ASCII prefix. Returns the decoded ASCII token or "".
func findUTF16JWT(data []byte, prefix string) string {
	if len(prefix) == 0 || len(data) < len(prefix)*2 {
		return ""
	}

	pattern := make([]byte, len(prefix)*2)
	for i := 0; i < len(prefix); i++ {
		pattern[i*2] = prefix[i]
		pattern[i*2+1] = 0
	}

	for i := 0; i <= len(data)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if data[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		var token strings.Builder
		for pos := i; pos+1 < len(data); pos += 2 {
			ch := data[pos]
			hi := data[pos+1]
			if hi != 0 || !isJWTChar(ch) {
				break
			}
			token.WriteByte(ch)
		}
		return token.String()
	}
	return ""
}

func isJWTChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.'
}
