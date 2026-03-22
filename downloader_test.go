package main

import "testing"

func TestValidatePDF_Valid(t *testing.T) {
	data := []byte("%PDF-1.4 fake pdf content padding to exceed 1KB" +
		string(make([]byte, 1024)))
	if err := validatePDF(data); err != nil {
		t.Fatalf("expected valid PDF, got error: %v", err)
	}
}

func TestValidatePDF_TooSmall(t *testing.T) {
	data := []byte("%PDF-1.4 tiny")
	if err := validatePDF(data); err == nil {
		t.Fatal("expected error for small file")
	}
}

func TestValidatePDF_BadMagic(t *testing.T) {
	data := make([]byte, 2048)
	copy(data, []byte("<html>Error</html>"))
	if err := validatePDF(data); err == nil {
		t.Fatal("expected error for non-PDF content")
	}
}
