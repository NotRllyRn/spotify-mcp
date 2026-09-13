package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtect(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		origin string
		want   int
	}{
		{"missing token", "", "", http.StatusUnauthorized},
		{"wrong token", "wrong", "", http.StatusUnauthorized},
		{"valid token", "secret", "", http.StatusNoContent},
		{"allowed origin", "secret", "https://example.com", http.StatusNoContent},
		{"denied origin", "secret", "https://evil.example", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			handler := protect(next, "secret", map[string]struct{}{"https://example.com": {}})
			req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			req.Header.Set("Origin", tt.origin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}
