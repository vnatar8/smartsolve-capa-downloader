package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueFilename_NoConflict(t *testing.T) {
	dir := t.TempDir()
	got := uniqueFilename(dir, "CAPA-2025-000043.pdf")
	expected := filepath.Join(dir, "CAPA-2025-000043.pdf")
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestUniqueFilename_WithConflict(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CAPA-2025-000043.pdf"), []byte("x"), 0644)

	got := uniqueFilename(dir, "CAPA-2025-000043.pdf")
	expected := filepath.Join(dir, "CAPA-2025-000043_1.pdf")
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestUniqueFilename_MultipleConflicts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CAPA-2025-000043.pdf"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "CAPA-2025-000043_1.pdf"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "CAPA-2025-000043_2.pdf"), []byte("x"), 0644)

	got := uniqueFilename(dir, "CAPA-2025-000043.pdf")
	expected := filepath.Join(dir, "CAPA-2025-000043_3.pdf")
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestSavePDF(t *testing.T) {
	dir := t.TempDir()
	data := []byte("%PDF-1.4 test content")

	path, err := savePDF(dir, "CAPA-2025-000043", data)
	if err != nil {
		t.Fatalf("savePDF failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read saved file: %v", err)
	}
	if string(content) != string(data) {
		t.Fatal("saved content does not match")
	}
}
