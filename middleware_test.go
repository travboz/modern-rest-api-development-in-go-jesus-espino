package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddCacheHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/lists", nil)
	rec := httptest.NewRecorder()

	handler := addCacheHeaders(testHandler)

	handler(rec, req)

	cacheHeaderWant := "public, max-age=300"

	if rec.Header().Get("Cache-Control") != cacheHeaderWant {
		t.Errorf("Not valid Cache-Control found, got %v, want %v", rec.Header().Get("Cache-Control"), "public, max-age=300")
	}

	if rec.Header().Get("Expires") == "" {
		t.Errorf("Not valid Expires, got %v, want not empty", rec.Header().Get("Expires"))
	}
}
