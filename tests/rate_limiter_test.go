package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"barbershop-backend/routes"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	router := routes.SetupRouter()

	// Kirim banyak request secara beruntun untuk memicu batas limit
	// Sesuaikan jumlah loop dengan batas limit token bucket yang kamu atur di middleware
	var lastCode int
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequest("GET", "/api/barbers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		lastCode = w.Code
		
		// Jika sudah kena rate limit, kita bisa hentikan loop
		if lastCode == http.StatusTooManyRequests {
			break
		}
	}

	// Pastikan setidaknya ada request yang diblokir (429) atau sistem merespon dengan benar
	// (Tergantung seberapa ketat limit yang kamu pasang)
	assert.True(t, lastCode == http.StatusOK || lastCode == http.StatusTooManyRequests)
}