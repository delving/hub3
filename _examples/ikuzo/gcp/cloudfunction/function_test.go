package function

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestF(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(F)
	handler.ServeHTTP(w, r)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Barbatos") {
		t.Errorf("wrong response body: got %v should contain %v", body, "Barbatos")
	}
}
