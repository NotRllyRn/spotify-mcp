package server

import (
	"context"
	"errors"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type searchInput struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
type topInput struct {
	TimeRange string `json:"time_range,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

func searchPage(in searchInput) (int, int, error) {
	if in.Query == "" {
		return 0, 0, errors.New("query is required")
	}
	return normalizedPage(pageInput{in.Limit, in.Offset}, 5, 10)
}

func registerSearchTools(server *mcp.Server, client *spotify.Client) {
	addTool(server, "search_tracks", "Search Spotify tracks.", func(ctx context.Context, in searchInput) (spotify.Paging[spotify.Track], error) {
		limit, offset, err := searchPage(in)
		if err != nil {
			return spotify.Paging[spotify.Track]{}, err
		}
		return client.SearchTracks(ctx, in.Query, limit, offset)
	})
	addTool(server, "search_albums", "Search Spotify albums.", func(ctx context.Context, in searchInput) (spotify.Paging[spotify.Album], error) {
		limit, offset, err := searchPage(in)
		if err != nil {
			return spotify.Paging[spotify.Album]{}, err
		}
		return client.SearchAlbums(ctx, in.Query, limit, offset)
	})
	addTool(server, "search_artists", "Search Spotify artists.", func(ctx context.Context, in searchInput) (spotify.Paging[spotify.Artist], error) {
		limit, offset, err := searchPage(in)
		if err != nil {
			return spotify.Paging[spotify.Artist]{}, err
		}
		return client.SearchArtists(ctx, in.Query, limit, offset)
	})
	addTool(server, "search_playlists", "Search Spotify playlists.", func(ctx context.Context, in searchInput) (spotify.Paging[spotify.Playlist], error) {
		limit, offset, err := searchPage(in)
		if err != nil {
			return spotify.Paging[spotify.Playlist]{}, err
		}
		return client.SearchPlaylists(ctx, in.Query, limit, offset)
	})
	addTool(server, "get_track", "Get one track by ID.", func(ctx context.Context, in idInput) (spotify.Track, error) {
		return client.GetTrack(ctx, in.ID)
	})
	addTool(server, "get_album", "Get one album by ID.", func(ctx context.Context, in idInput) (spotify.Album, error) {
		return client.GetAlbum(ctx, in.ID)
	})
	addTool(server, "get_artist", "Get one artist by ID.", func(ctx context.Context, in idInput) (spotify.Artist, error) {
		return client.GetArtist(ctx, in.ID)
	})
	addTool(server, "get_artist_albums", "Get albums by an artist.", func(ctx context.Context, in idPageInput) (spotify.Paging[spotify.Album], error) {
		limit, offset, err := normalizedPage(pageInput{in.Limit, in.Offset}, 20, 50)
		if err != nil {
			return spotify.Paging[spotify.Album]{}, err
		}
		return client.GetArtistAlbums(ctx, in.ID, limit, offset)
	})
	addTool(server, "get_playlist", "Get a playlist by ID.", func(ctx context.Context, in idInput) (spotify.Playlist, error) {
		return client.GetPlaylist(ctx, in.ID)
	})
	addTool(server, "get_playlist_items", "Get available items in a playlist.", func(ctx context.Context, in idPageInput) (spotify.Paging[spotify.PlaylistItem], error) {
		limit, offset, err := normalizedPage(pageInput{in.Limit, in.Offset}, 20, 50)
		if err != nil {
			return spotify.Paging[spotify.PlaylistItem]{}, err
		}
		return client.GetPlaylistItems(ctx, in.ID, limit, offset)
	})
	addTool(server, "get_recently_played", "Get recently played tracks.", func(ctx context.Context, in pageInput) (any, error) {
		limit, _, err := normalizedPage(in, 20, 50)
		if err != nil {
			return nil, err
		}
		return client.GetRecentlyPlayed(ctx, limit)
	})
	addTool(server, "get_top_tracks", "Get the user's top tracks.", func(ctx context.Context, in topInput) (spotify.Paging[spotify.Track], error) {
		limit, offset, timeRange, err := normalizeTop(in)
		if err != nil {
			return spotify.Paging[spotify.Track]{}, err
		}
		return client.TopTracks(ctx, timeRange, limit, offset)
	})
	addTool(server, "get_top_artists", "Get the user's top artists.", func(ctx context.Context, in topInput) (spotify.Paging[spotify.Artist], error) {
		limit, offset, timeRange, err := normalizeTop(in)
		if err != nil {
			return spotify.Paging[spotify.Artist]{}, err
		}
		return client.TopArtists(ctx, timeRange, limit, offset)
	})
	addTool(server, "get_my_profile", "Get the current Spotify profile.", func(ctx context.Context, _ emptyInput) (spotify.Profile, error) {
		return client.Profile(ctx)
	})
}

func normalizeTop(in topInput) (int, int, string, error) {
	limit, offset, err := normalizedPage(pageInput{in.Limit, in.Offset}, 20, 50)
	if in.TimeRange == "" {
		in.TimeRange = "medium_term"
	}
	return limit, offset, in.TimeRange, err
}
