package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// uniqueFilename returns a filepath that doesn't conflict with existing files.
// If "name.pdf" exists, tries "name_1.pdf", "name_2.pdf", etc.
func uniqueFilename(dir string, filename string) string {
	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fullPath
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// savePDF saves PDF data to the output directory with duplicate handling.
// Returns the final filepath used.
func savePDF(outputDir string, capaNumber string, data []byte) (string, error) {
	filename := capaNumber + ".pdf"
	destPath := uniqueFilename(outputDir, filename)

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", destPath, err)
	}
	return destPath, nil
}
