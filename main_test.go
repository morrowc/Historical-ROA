package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

	body := rec.Body.String()
	if !strings.Contains(body, "Forbidden") {
		t.Errorf("Expected body to contain 'Forbidden', got %q", body)
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
