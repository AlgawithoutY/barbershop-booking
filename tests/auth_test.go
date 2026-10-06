package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"barbershop-backend/routes"

	"github.com/stretchr/testify/assert"
)

func TestProtectedEndpointWithoutToken(t *testing.T) {
	router := routes.SetupRouter()

	// Mencoba mengakses endpoint booking tanpa menyertakan Header Authorization Bearer Token
	req, _ := http.NewRequest("GET", "/api/my-bookings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Harus ditolak dengan status 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}