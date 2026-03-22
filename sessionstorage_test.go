package main

import (
	"os"
	"testing"
)

func TestChromeLocalStoragePath(t *testing.T) {
	path := chromeLocalStoragePath()
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("Chrome Local Storage not found: %v", err)
	}
}

func TestReadJWTFromLocalStorage(t *testing.T) {
	token, err := readJWTFromLocalStorage()
	if err != nil {
		t.Skipf("Could not read JWT (may need SmartSolve open in Chrome): %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if len(token) < 50 {
		t.Fatalf("token seems too short: %s", token)
	}
	t.Logf("Found token: %s...%s", token[:20], token[len(token)-10:])
}

func TestExtractJWT(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bare JWT",
			input: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig",
			want:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig",
		},
		{
			name:  "JWT with surrounding quotes",
			input: `"eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxIn0.abc"`,
			want:  "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiIxIn0.abc",
		},
		{
			name:  "JWT with prefix text",
			input: "Bearer eyJhbGciOiJSUzI1NiJ9.payload.sig more stuff",
			want:  "eyJhbGciOiJSUzI1NiJ9.payload.sig",
		},
		{
			name:  "no JWT",
			input: "just a regular string",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJWT(tt.input)
			if got != tt.want {
				t.Errorf("extractJWT(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
