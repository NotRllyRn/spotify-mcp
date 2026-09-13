package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChallenge(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	if got, want := Challenge(verifier), "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"; got != want {
		t.Fatalf("Challenge() = %q, want %q", got, want)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token.json")
	store := Store{Path: path}
	want := State{RefreshToken: "refresh", AuthorizedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestTokenRefreshRotation(t *testing.T) {
	var form url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ = url.ParseQuery(string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"access","refresh_token":"rotated","expires_in":3600}`)
	}))
	defer server.Close()
	store := Store{Path: filepath.Join(t.TempDir(), "token.json")}
	if err := store.Save(State{RefreshToken: "original", AuthorizedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	manager := NewTokenManager(server.Client(), store, "client", server.URL)
	got, err := manager.Token(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "access" || form.Get("refresh_token") != "original" || form.Get("client_id") != "client" {
		t.Fatalf("unexpected token exchange: token=%q form=%v", got, form)
	}
	state, err := store.Load()
	if err != nil || state.RefreshToken != "rotated" {
		t.Fatalf("rotated state = %#v, error = %v", state, err)
	}
}

func TestExpiredAuthorizationRequiresAuthorization(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "token.json")}
	_ = store.Save(State{RefreshToken: "expired", AuthorizedAt: time.Now().AddDate(0, -6, -1)})
	manager := NewTokenManager(http.DefaultClient, store, "client", "http://unused")
	if _, err := manager.Token(context.Background(), false); err != ErrReauthorizationRequired {
		t.Fatalf("error = %v, want %v", err, ErrReauthorizationRequired)
	}
}

func TestInvalidGrantRequiresAuthorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
	}))
	defer server.Close()
	store := Store{Path: filepath.Join(t.TempDir(), "token.json")}
	_ = store.Save(State{RefreshToken: "expired", AuthorizedAt: time.Now()})
	manager := NewTokenManager(server.Client(), store, "client", server.URL)
	if _, err := manager.Token(context.Background(), false); err != ErrReauthorizationRequired {
		t.Fatalf("error = %v, want %v", err, ErrReauthorizationRequired)
	}
}
