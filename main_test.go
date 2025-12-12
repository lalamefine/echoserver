package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSCTHandlerNotFound(t *testing.T) {
	old := sctToken
	defer func() { sctToken = old }()
	sctToken = ""

	req := httptest.NewRequest(http.MethodGet, "/.well-known/scale-test-claim-token.txt", nil)
	r := httptest.NewRecorder()

	sctHandler(r, req)

	if r.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when token not set, got %d", r.Code)
	}
}

func TestSCTHandlerServesToken(t *testing.T) {
	old := sctToken
	defer func() { sctToken = old }()
	sctToken = "my-test-token-123"

	req := httptest.NewRequest(http.MethodGet, "/.well-known/scale-test-claim-token.txt", nil)
	r := httptest.NewRecorder()

	sctHandler(r, req)

	if r.Code != http.StatusOK {
		t.Fatalf("expected 200 when token set, got %d", r.Code)
	}

	if got := r.Body.String(); got != sctToken {
		t.Fatalf("unexpected body: %q", got)
	}
}
