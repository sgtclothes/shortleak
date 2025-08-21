package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"shortleak/server"
)

func TestCORS(t *testing.T) {
	router := server.SetupRouter()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("expected CORS allow-origin http://localhost:5173, got %s",
			w.Header().Get("Access-Control-Allow-Origin"))
	}
}
