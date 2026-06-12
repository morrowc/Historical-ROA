package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestErrorHandlerEscaping(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	unsafeAlert := "<script>alert('xss')</script>"
	ErrorHandler(rec, req, http.StatusInternalServerError, unsafeAlert, fmt.Errorf("test error"))

	body := rec.Body.String()
	if strings.Contains(body, unsafeAlert) {
		t.Errorf("Response body contains unescaped alert: %q", body)
	}

	escapedAlert := html.EscapeString(unsafeAlert)
	if !strings.Contains(body, escapedAlert) {
		t.Errorf("Response body does not contain escaped alert: %q", body)
	}
}

func TestPullToDB_MissingCronHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/update", nil)

	// Should fail with Forbidden and NOT panic because it exits before BQ calls.
	pullToDB(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %v, got %v", http.StatusForbidden, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to start with 'text/plain', got %q", contentType)
	}

	body := rec.Body.String()
	expectedBody := "Error 403: Forbidden: OIDC verification failed: missing Authorization header\n"
	if body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}

func TestPullToDB_WrongCronHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/update", nil)
	req.Header.Set("X-Appengine-Cron", "false")

	pullToDB(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %v, got %v", http.StatusForbidden, rec.Code)
	}
}

func TestTextErrorHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	TextErrorHandler(rec, http.StatusBadRequest, "Bad Request occurred", fmt.Errorf("some detail"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %v, got %v", http.StatusBadRequest, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to start with 'text/plain', got %q", contentType)
	}

	body := rec.Body.String()
	expectedBody := "Error 400: Bad Request occurred: some detail\n"
	if body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}

func TestNormalizeASN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"15169", "AS15169", false},
		{"AS15169", "AS15169", false},
		{"as15169", "AS15169", false},
		{"", "", false},
		{"   15169   ", "AS15169", false},
		{"AS", "", true},
		{"15169foo", "", true},
		{"ASfoobar", "", true},
	}

	for _, tc := range tests {
		got, err := normalizeASN(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("normalizeASN(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if got != tc.expected {
			t.Errorf("normalizeASN(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestNormalizePrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"8.8.8.0/24", "8.8.8.0/24", false},
		{"1.1.1.1", "1.1.1.0/24", false},
		{"2001:4860:4860::8888", "2001:4860:4860::/48", false},
		{"2001:db8::/32", "2001:db8::/32", false},
		{"", "", false},
		{"   8.8.8.8   ", "8.8.8.0/24", false},
		{"invalid-ip", "", true},
		{"8.8.8.0/99", "", true},
	}

	for _, tc := range tests {
		got, err := normalizePrefix(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("normalizePrefix(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			continue
		}
		if got != tc.expected {
			t.Errorf("normalizePrefix(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestComputeAvailabilityRanges(t *testing.T) {
	// Consecutive daily updates
	d1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	d2 := d1.Add(24 * time.Hour)
	d3 := d2.Add(24 * time.Hour)

	// Large gap (interruption)
	d4 := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	d5 := d4.Add(24 * time.Hour)

	times := []time.Time{d1, d2, d3, d4, d5}

	got := computeAvailabilityRanges(times, 26*time.Hour)
	expected := []string{
		"Jan 1 2026 -> Jan 3 2026",
		"Feb 1 2026 -> Feb 2 2026",
	}

	if len(got) != len(expected) {
		t.Fatalf("got %d ranges, want %d", len(got), len(expected))
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Errorf("range %d = %q, want %q", i, got[i], expected[i])
		}
	}
}


