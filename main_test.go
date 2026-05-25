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

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to start with 'text/plain', got %q", contentType)
	}

	body := rec.Body.String()
	expectedBody := "Error 403: Forbidden: missing X-Appengine-Cron or X-CloudScheduler header\n"
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

