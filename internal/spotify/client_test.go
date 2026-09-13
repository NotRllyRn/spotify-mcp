package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeTokens struct{ forces []bool }

func (f *fakeTokens) Token(_ context.Context, force bool) (string, error) {
	f.forces = append(f.forces, force)
	if force {
		return "refreshed", nil
	}
	return "initial", nil
}

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *fakeTokens, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	tokens := new(fakeTokens)
	client, err := NewClient(server.Client(), tokens, server.URL+"/v1/")
	if err != nil {
		t.Fatal(err)
	}
	return client, tokens, server.Close
}

func TestDoResponses(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated, http.StatusNoContent} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client, _, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/test" || r.Header.Get("Authorization") != "Bearer initial" {
					t.Errorf("unexpected request: %s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
				}
				w.WriteHeader(status)
				if status != http.StatusNoContent {
					_, _ = io.WriteString(w, `{"id":"ok"}`)
				}
			})
			defer closeServer()
			var result struct {
				ID string `json:"id"`
			}
			if err := client.Do(context.Background(), http.MethodGet, "test", nil, nil, &result); err != nil {
				t.Fatal(err)
			}
			if status != http.StatusNoContent && result.ID != "ok" {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestDoRefreshesOnce(t *testing.T) {
	requests := 0
	client, tokens, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != "Bearer refreshed" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer closeServer()
	if err := client.Do(context.Background(), http.MethodPut, "test", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(tokens.forces) != 2 || tokens.forces[0] || !tokens.forces[1] {
		t.Fatalf("force calls = %v", tokens.forces)
	}
}

func TestDoSecondUnauthorizedFails(t *testing.T) {
	client, _, closeServer := testClient(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	defer closeServer()
	err := client.Do(context.Background(), http.MethodGet, "test", nil, nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "unauthorized" {
		t.Fatalf("error = %v", err)
	}
}

func TestDoRateLimit(t *testing.T) {
	requests := 0
	client, _, closeServer := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Retry-After", "0")
		if requests == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer closeServer()
	if err := client.Do(context.Background(), http.MethodGet, "test", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestDoRejectsLongRateLimit(t *testing.T) {
	client, _, closeServer := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "31")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer closeServer()
	var apiErr *APIError
	if err := client.Do(context.Background(), http.MethodGet, "test", nil, nil, nil); !errors.As(err, &apiErr) || apiErr.Code != "rate_limited" {
		t.Fatalf("error = %v", err)
	}
}

func TestCurrentMutationPathsAndChunks(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	client, _, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		w.WriteHeader(http.StatusNoContent)
	})
	defer closeServer()
	uris := make([]string, 41)
	for i := range uris {
		uris[i] = "spotify:track:x"
	}
	if _, err := client.SaveToLibrary(context.Background(), uris); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RemovePlaylistItems(context.Background(), "playlist", uris[:1]); err != nil {
		t.Fatal(err)
	}
	want := []string{"PUT /v1/me/library", "PUT /v1/me/library", "DELETE /v1/playlists/playlist/items"}
	if strings.Join(paths, ",") != strings.Join(want, ",") {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	if got := len(bodies[0]["uris"].([]any)); got != 40 {
		t.Fatalf("first library chunk = %d, want 40", got)
	}
	if _, ok := bodies[2]["items"]; !ok {
		t.Fatalf("playlist removal body = %v", bodies[2])
	}
}
