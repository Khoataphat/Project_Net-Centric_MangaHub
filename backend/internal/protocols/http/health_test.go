package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	// Switch to test mode so you don't get such noisy output
	gin.SetMode(gin.TestMode)

	// Setup router
	r := gin.Default()
	r.GET("/health", HealthHandler)

	// Create a request to send to the above route
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check if the response was what we expected
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	expected := `{"message":"MangaHub Backend is running","status":"up"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, but got %s", expected, w.Body.String())
	}
}
