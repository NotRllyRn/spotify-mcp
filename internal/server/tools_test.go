package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type staticToken struct{}

func (staticToken) Token(context.Context, bool) (string, error) { return "token", nil }

func TestToolCatalogAndCalls(t *testing.T) {
	requests := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path == "/v1/me/player/volume" && r.URL.Query().Get("volume_percent") != "25" {
			t.Errorf("volume query = %v", r.URL.Query())
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer api.Close()
	spotifyClient, err := spotify.NewClient(api.Client(), staticToken{}, api.URL+"/v1/")
	if err != nil {
		t.Fatal(err)
	}
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	RegisterTools(mcpServer, spotifyClient)
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := mcpServer.Connect(context.Background(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	tools, err := clientSession.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 38 {
		t.Fatalf("tool count = %d, want 38", len(tools.Tools))
	}
	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "set_volume", Arguments: map[string]any{"volume_percent": 25}})
	if err != nil || result.IsError {
		t.Fatalf("set_volume result = %#v, error = %v", result, err)
	}
	result, err = clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "search_tracks", Arguments: map[string]any{"query": "x", "limit": 11}})
	if err != nil || !result.IsError {
		t.Fatalf("invalid search result = %#v, error = %v", result, err)
	}
	if requests != 1 {
		t.Fatalf("Spotify requests = %d, want 1", requests)
	}
	if text, ok := result.Content[0].(*mcp.TextContent); !ok || len(text.Text) == 0 {
		t.Fatalf("error content = %#v", result.Content)
	}
}
