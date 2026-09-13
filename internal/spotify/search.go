package spotify

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

func (c *Client) search(ctx context.Context, kind, query string, limit, offset int, out any) error {
	if query == "" {
		return errors.New("query is required")
	}
	values, err := pagination(limit, offset, 10)
	if err != nil {
		return err
	}
	values.Set("q", query)
	values.Set("type", kind)
	return c.Do(ctx, http.MethodGet, "search", values, nil, out)
}

func (c *Client) SearchTracks(ctx context.Context, query string, limit, offset int) (Paging[Track], error) {
	var response struct {
		Tracks Paging[Track] `json:"tracks"`
	}
	err := c.search(ctx, "track", query, limit, offset, &response)
	return response.Tracks, err
}

func (c *Client) SearchAlbums(ctx context.Context, query string, limit, offset int) (Paging[Album], error) {
	var response struct {
		Albums Paging[Album] `json:"albums"`
	}
	err := c.search(ctx, "album", query, limit, offset, &response)
	return response.Albums, err
}

func (c *Client) SearchArtists(ctx context.Context, query string, limit, offset int) (Paging[Artist], error) {
	var response struct {
		Artists Paging[Artist] `json:"artists"`
	}
	err := c.search(ctx, "artist", query, limit, offset, &response)
	return response.Artists, err
}

func (c *Client) SearchPlaylists(ctx context.Context, query string, limit, offset int) (Paging[Playlist], error) {
	var response struct {
		Playlists Paging[Playlist] `json:"playlists"`
	}
	err := c.search(ctx, "playlist", query, limit, offset, &response)
	return response.Playlists, err
}

func (c *Client) GetTrack(ctx context.Context, id string) (Track, error) {
	var result Track
	if err := requireID("id", id); err != nil {
		return result, err
	}
	err := c.Do(ctx, http.MethodGet, "tracks/"+id, nil, nil, &result)
	return result, err
}

func (c *Client) GetAlbum(ctx context.Context, id string) (Album, error) {
	var result Album
	if err := requireID("id", id); err != nil {
		return result, err
	}
	err := c.Do(ctx, http.MethodGet, "albums/"+id, nil, nil, &result)
	return result, err
}

func (c *Client) GetArtist(ctx context.Context, id string) (Artist, error) {
	var result Artist
	if err := requireID("id", id); err != nil {
		return result, err
	}
	err := c.Do(ctx, http.MethodGet, "artists/"+id, nil, nil, &result)
	return result, err
}

func (c *Client) GetArtistAlbums(ctx context.Context, id string, limit, offset int) (Paging[Album], error) {
	var result Paging[Album]
	if err := requireID("id", id); err != nil {
		return result, err
	}
	query, err := pagination(limit, offset, 50)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "artists/"+id+"/albums", query, nil, &result)
	return result, err
}

func (c *Client) GetPlaylist(ctx context.Context, id string) (Playlist, error) {
	var result Playlist
	if err := requireID("id", id); err != nil {
		return result, err
	}
	err := c.Do(ctx, http.MethodGet, "playlists/"+id, nil, nil, &result)
	return result, err
}

func (c *Client) GetPlaylistItems(ctx context.Context, id string, limit, offset int) (Paging[PlaylistItem], error) {
	var result Paging[PlaylistItem]
	if err := requireID("id", id); err != nil {
		return result, err
	}
	query, err := pagination(limit, offset, 50)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "playlists/"+id+"/items", query, nil, &result)
	return result, err
}

func (c *Client) GetRecentlyPlayed(ctx context.Context, limit int) (Paging[struct {
	Track    Track  `json:"track"`
	PlayedAt string `json:"played_at"`
}], error) {
	var result Paging[struct {
		Track    Track  `json:"track"`
		PlayedAt string `json:"played_at"`
	}]
	query, err := pagination(limit, 0, 50)
	if err != nil {
		return result, err
	}
	query.Del("offset")
	err = c.Do(ctx, http.MethodGet, "me/player/recently-played", query, nil, &result)
	return result, err
}

func (c *Client) TopTracks(ctx context.Context, timeRange string, limit, offset int) (Paging[Track], error) {
	var result Paging[Track]
	query, err := topQuery(timeRange, limit, offset)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "me/top/tracks", query, nil, &result)
	return result, err
}

func (c *Client) TopArtists(ctx context.Context, timeRange string, limit, offset int) (Paging[Artist], error) {
	var result Paging[Artist]
	query, err := topQuery(timeRange, limit, offset)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "me/top/artists", query, nil, &result)
	return result, err
}

func topQuery(timeRange string, limit, offset int) (url.Values, error) {
	if timeRange != "short_term" && timeRange != "medium_term" && timeRange != "long_term" {
		return nil, errors.New("time_range must be short_term, medium_term, or long_term")
	}
	query, err := pagination(limit, offset, 50)
	if err == nil {
		query.Set("time_range", timeRange)
	}
	return query, err
}

func (c *Client) Profile(ctx context.Context) (Profile, error) {
	var result Profile
	err := c.Do(ctx, http.MethodGet, "me", nil, nil, &result)
	return result, err
}
