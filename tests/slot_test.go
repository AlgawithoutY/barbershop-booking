package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"barbershop-backend/config"
	"barbershop-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Load file .env dari root project (naik satu folder dari folder tests)
	_ = godotenv.Load("../.env")
	
	// Fallback menggunakan konfigurasi database kamu yang sebenarnya
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "127.0.0.1")
		os.Setenv("DB_PORT", "3306")
		os.Setenv("DB_USER", "root")
		os.Setenv("DB_PASSWORD", "Alga")
		os.Setenv("DB_NAME", "barbershop_db")
	}

	config.ConnectDatabase()
}

func TestGetSlotsEndpoint(t *testing.T) {
	// Set Gin ke Mode Test agar Rate Limiter mengabaikan limitasi saat testing
	gin.SetMode(gin.TestMode)

	router := routes.SetupRouter()

	req, _ := http.NewRequest("GET", "/api/barbers/1/slots?date=2026-09-06", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}