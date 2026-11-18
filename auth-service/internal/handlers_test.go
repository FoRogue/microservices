package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRejectsInvalidInput(t *testing.T) {
	handler := NewHandler(NewService(nil, "test-secret")).Routes()
	request := httptest.NewRequest(http.MethodPost, "/api/register",
		strings.NewReader(`{"username":"  ","password":"password"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestProfileRejectsInvalidToken(t *testing.T) {
	handler := NewHandler(NewService(nil, "test-secret")).Routes()
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
