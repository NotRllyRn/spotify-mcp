package server

import (
	"context"
	"errors"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createPlaylistInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public,omitempty"`
}
type updatePlaylistInput struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Public      *bool   `json:"public,omitempty"`
}
type playlistItemsInput struct {
	ID       string   `json:"id"`
	URIs     []string `json:"uris"`
	Position *int     `json:"position,omitempty"`
}
type reorderInput struct {
	ID           string `json:"id"`
	RangeStart   int    `json:"range_start"`
	InsertBefore int    `json:"insert_before"`
	RangeLength  *int   `json:"range_length,omitempty"`
}

func registerPlaylistTools(server *mcp.Server, client *spotify.Client) {
	addTool(server, "list_my_playlists", "List the user's playlists.", func(ctx context.Context, in pageInput) (spotify.Paging[spotify.Playlist], error) {
		limit, offset, err := normalizedPage(in, 20, 50)
		if err != nil {
			return spotify.Paging[spotify.Playlist]{}, err
		}
		return client.ListMyPlaylists(ctx, limit, offset)
	})
	addTool(server, "create_playlist", "Create a playlist.", func(ctx context.Context, in createPlaylistInput) (spotify.Playlist, error) {
		return client.CreatePlaylist(ctx, in.Name, in.Description, in.Public)
	})
	addTool(server, "update_playlist", "Update playlist details.", func(ctx context.Context, in updatePlaylistInput) (acknowledgment, error) {
		body := make(map[string]any)
		if in.Name != nil {
			body["name"] = *in.Name
		}
		if in.Description != nil {
			body["description"] = *in.Description
		}
		if in.Public != nil {
			body["public"] = *in.Public
		}
		return acknowledgment{true}, client.UpdatePlaylist(ctx, in.ID, body)
	})
	addTool(server, "add_playlist_items", "Add Spotify URIs to a playlist.", func(ctx context.Context, in playlistItemsInput) (spotify.MutationResult, error) {
		return client.AddPlaylistItems(ctx, in.ID, in.URIs, in.Position)
	})
	addTool(server, "remove_playlist_items", "Remove Spotify URIs from a playlist.", func(ctx context.Context, in playlistItemsInput) (spotify.MutationResult, error) {
		if in.Position != nil {
			return spotify.MutationResult{}, errors.New("position is not valid when removing items")
		}
		return client.RemovePlaylistItems(ctx, in.ID, in.URIs)
	})
	addTool(server, "reorder_playlist_items", "Reorder items in a playlist.", func(ctx context.Context, in reorderInput) (acknowledgment, error) {
		return acknowledgment{true}, client.ReorderPlaylistItems(ctx, in.ID, in.RangeStart, in.InsertBefore, in.RangeLength)
	})
}
