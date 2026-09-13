package spotify

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const libraryBatchSize = 40

type SavedTrack struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

type SavedAlbum struct {
	AddedAt string `json:"added_at"`
	Album   Album  `json:"album"`
}

func (c *Client) SavedTracks(ctx context.Context, limit, offset int) (Paging[SavedTrack], error) {
	var result Paging[SavedTrack]
	query, err := pagination(limit, offset, 50)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "me/tracks", query, nil, &result)
	return result, err
}

func (c *Client) SavedAlbums(ctx context.Context, limit, offset int) (Paging[SavedAlbum], error) {
	var result Paging[SavedAlbum]
	query, err := pagination(limit, offset, 50)
	if err != nil {
		return result, err
	}
	err = c.Do(ctx, http.MethodGet, "me/albums", query, nil, &result)
	return result, err
}

func (c *Client) SaveToLibrary(ctx context.Context, uris []string) (MutationResult, error) {
	return c.mutateLibrary(ctx, http.MethodPut, uris)
}

func (c *Client) RemoveFromLibrary(ctx context.Context, uris []string) (MutationResult, error) {
	return c.mutateLibrary(ctx, http.MethodDelete, uris)
}

func (c *Client) mutateLibrary(ctx context.Context, method string, uris []string) (MutationResult, error) {
	if len(uris) == 0 {
		return MutationResult{}, errors.New("uris must not be empty")
	}
	result := MutationResult{}
	for start := 0; start < len(uris); start += libraryBatchSize {
		end := min(start+libraryBatchSize, len(uris))
		if err := c.Do(ctx, method, "me/library", nil, map[string]any{"uris": uris[start:end]}, nil); err != nil {
			return result, err
		}
		result.Completed += end - start
	}
	return result, nil
}

func (c *Client) CheckLibrary(ctx context.Context, uris []string) ([]bool, error) {
	if len(uris) == 0 {
		return nil, errors.New("uris must not be empty")
	}
	result := make([]bool, 0, len(uris))
	for start := 0; start < len(uris); start += libraryBatchSize {
		end := min(start+libraryBatchSize, len(uris))
		var chunk []bool
		query := url.Values{"uris": {strings.Join(uris[start:end], ",")}}
		if err := c.Do(ctx, http.MethodGet, "me/library/contains", query, nil, &chunk); err != nil {
			return result, err
		}
		result = append(result, chunk...)
	}
	return result, nil
}
