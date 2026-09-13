package spotify

import (
	"context"
	"errors"
	"net/http"
)

const playlistBatchSize = 100

func (c *Client) ListMyPlaylists(ctx context.Context, limit, offset int) (Paging[Playlist], error) {
	var result Paging[Playlist]
	query, err := pagination(limit, offset, 50)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "me/playlists", query, nil, &result)
	return result, err
}

func (c *Client) CreatePlaylist(ctx context.Context, name, description string, public bool) (Playlist, error) {
	var result Playlist
	if name == "" {
		return result, errors.New("name is required")
	}
	body := map[string]any{"name": name, "description": description, "public": public}
	err := c.Do(ctx, http.MethodPost, "me/playlists", nil, body, &result)
	return result, err
}

func (c *Client) UpdatePlaylist(ctx context.Context, id string, body map[string]any) error {
	if err := requireID("id", id); err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("at least one field must be supplied")
	}
	return c.Do(ctx, http.MethodPut, "playlists/"+id, nil, body, nil)
}

func (c *Client) AddPlaylistItems(ctx context.Context, id string, uris []string, position *int) (MutationResult, error) {
	if err := requireID("id", id); err != nil {
		return MutationResult{}, err
	}
	if len(uris) == 0 {
		return MutationResult{}, errors.New("uris must not be empty")
	}
	if position != nil && *position < 0 {
		return MutationResult{}, errors.New("position must be at least 0")
	}
	result := MutationResult{}
	for start := 0; start < len(uris); start += playlistBatchSize {
		end := min(start+playlistBatchSize, len(uris))
		body := map[string]any{"uris": uris[start:end]}
		if position != nil {
			body["position"] = *position + start
		}
		if err := c.Do(ctx, http.MethodPost, "playlists/"+id+"/items", nil, body, nil); err != nil {
			return result, err
		}
		result.Completed += end - start
	}
	return result, nil
}

func (c *Client) RemovePlaylistItems(ctx context.Context, id string, uris []string) (MutationResult, error) {
	if err := requireID("id", id); err != nil {
		return MutationResult{}, err
	}
	if len(uris) == 0 {
		return MutationResult{}, errors.New("uris must not be empty")
	}
	result := MutationResult{}
	for start := 0; start < len(uris); start += playlistBatchSize {
		end := min(start+playlistBatchSize, len(uris))
		items := make([]map[string]string, end-start)
		for i, uri := range uris[start:end] {
			items[i] = map[string]string{"uri": uri}
		}
		if err := c.Do(ctx, http.MethodDelete, "playlists/"+id+"/items", nil, map[string]any{"items": items}, nil); err != nil {
			return result, err
		}
		result.Completed += end - start
	}
	return result, nil
}

func (c *Client) ReorderPlaylistItems(ctx context.Context, id string, rangeStart, insertBefore int, rangeLength *int) error {
	if err := requireID("id", id); err != nil {
		return err
	}
	if rangeStart < 0 || insertBefore < 0 || (rangeLength != nil && *rangeLength < 1) {
		return errors.New("positions must be non-negative and range_length must be positive")
	}
	body := map[string]any{"range_start": rangeStart, "insert_before": insertBefore}
	if rangeLength != nil {
		body["range_length"] = *rangeLength
	}
	return c.Do(ctx, http.MethodPut, "playlists/"+id+"/items", nil, body, nil)
}
