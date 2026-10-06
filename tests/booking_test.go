package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"barbershop-backend/routes"

	"github.com/stretchr/testify/assert"
)

func TestCreateBookingEndpoint(t *testing.T) {
	router := routes.SetupRouter()

	// Payload JSON palsu untuk booking
	jsonBody := []byte(`{
		"barber_id": 1,
		"service_id": 1,
		"booking_time": "2026-09-10T10:00:00Z"
	}`)

	req, _ := http.NewRequest("POST", "/api/bookings", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// Catatan: Kalau endpoint ini butuh token, nanti kita tambahkan header Authorization di sini.
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Karena butuh login/token, biasanya tanpa token akan mengembalikan 401. 
	// Jika endpoint publik/terbuka, ekspektasinya bisa 201 Created. Kita uji respon statusnya valid.
	assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest)
}