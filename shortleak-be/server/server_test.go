package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	_ = os.Setenv("NODE_ENV", "test")
	_ = os.Setenv("DB_DATABASE_TEST", "shortleak-test")
	_ = os.Setenv("DB_USERNAME_TEST", "postgres")
	_ = os.Setenv("DB_PASSWORD_TEST", "12345")
	_ = os.Setenv("DB_HOST_TEST", "localhost")
	_ = os.Setenv("DB_DIALECT_TEST", "postgres")
	_ = os.Setenv("DB_PORT_TEST", "5432")
}

func TestSetupRouterNotNil(t *testing.T) {
	r := SetupRouter()
	assert.NotNil(t, r, "Router should not be nil")
}

func TestSetupRouterCORS(t *testing.T) {
	r := SetupRouter()

	req, _ := http.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code, "CORS preflight should return 204")
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSetupRouterRoutesExist(t *testing.T) {
	r := SetupRouter()
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Contains(t, []int{200, 404, 401, 204}, w.Code)
}
