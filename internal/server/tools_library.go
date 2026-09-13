package server

import (
	"context"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type urisInput struct {
	URIs []string `json:"uris"`
}
type libraryCheckOutput struct {
	URIs  []string `json:"uris"`
	Saved []bool   `json:"saved"`
}

func registerLibraryTools(server *mcp.Server, client *spotify.Client) {
	addTool(server, "get_saved_tracks", "Get saved tracks.", func(ctx context.Context, in pageInput) (spotify.Paging[spotify.SavedTrack], error) {
		limit, offset, err := normalizedPage(in, 20, 50)
		if err != nil {
			return spotify.Paging[spotify.SavedTrack]{}, err
		}
		return client.SavedTracks(ctx, limit, offset)
	})
	addTool(server, "get_saved_albums", "Get saved albums.", func(ctx context.Context, in pageInput) (spotify.Paging[spotify.SavedAlbum], error) {
		limit, offset, err := normalizedPage(in, 20, 50)
		if err != nil {
			return spotify.Paging[spotify.SavedAlbum]{}, err
		}
		return client.SavedAlbums(ctx, limit, offset)
	})
	addTool(server, "save_to_library", "Save Spotify URIs to the library.", func(ctx context.Context, in urisInput) (spotify.MutationResult, error) {
		return client.SaveToLibrary(ctx, in.URIs)
	})
	addTool(server, "remove_from_library", "Remove Spotify URIs from the library.", func(ctx context.Context, in urisInput) (spotify.MutationResult, error) {
		return client.RemoveFromLibrary(ctx, in.URIs)
	})
	addTool(server, "check_library", "Check whether Spotify URIs are saved.", func(ctx context.Context, in urisInput) (libraryCheckOutput, error) {
		saved, err := client.CheckLibrary(ctx, in.URIs)
		return libraryCheckOutput{URIs: in.URIs, Saved: saved}, err
	})
}
