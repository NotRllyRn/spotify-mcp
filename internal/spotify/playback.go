package spotify

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

func deviceQuery(deviceID string) url.Values {
	if deviceID == "" {
		return nil
	}
	return url.Values{"device_id": {deviceID}}
}

func (c *Client) PlaybackState(ctx context.Context) (PlaybackState, error) {
	var result PlaybackState
	err := c.Do(ctx, http.MethodGet, "me/player", nil, nil, &result)
	return result, err
}

func (c *Client) Play(ctx context.Context, deviceID, contextURI string, uris []string, offsetPosition int, positionMS *int) error {
	if contextURI != "" && len(uris) > 0 {
		return errors.New("context_uri and uris cannot both be supplied")
	}
	if offsetPosition < 0 {
		return errors.New("offset_position must be at least 0")
	}
	if positionMS != nil && *positionMS < 0 {
		return errors.New("position_ms must be at least 0")
	}
	body := make(map[string]any)
	if contextURI != "" {
		body["context_uri"] = contextURI
		body["offset"] = map[string]int{"position": offsetPosition}
	} else if len(uris) > 0 {
		body["uris"] = uris
	}
	if positionMS != nil {
		body["position_ms"] = *positionMS
	}
	if len(body) == 0 {
		body = nil
	}
	return c.Do(ctx, http.MethodPut, "me/player/play", deviceQuery(deviceID), body, nil)
}

func (c *Client) Pause(ctx context.Context, deviceID string) error {
	return c.Do(ctx, http.MethodPut, "me/player/pause", deviceQuery(deviceID), nil, nil)
}

func (c *Client) Next(ctx context.Context, deviceID string) error {
	return c.Do(ctx, http.MethodPost, "me/player/next", deviceQuery(deviceID), nil, nil)
}

func (c *Client) Previous(ctx context.Context, deviceID string) error {
	return c.Do(ctx, http.MethodPost, "me/player/previous", deviceQuery(deviceID), nil, nil)
}

func (c *Client) Seek(ctx context.Context, positionMS int, deviceID string) error {
	if positionMS < 0 {
		return errors.New("position_ms must be at least 0")
	}
	query := deviceQuery(deviceID)
	if query == nil {
		query = make(url.Values)
	}
	query.Set("position_ms", strconv.Itoa(positionMS))
	return c.Do(ctx, http.MethodPut, "me/player/seek", query, nil, nil)
}

func (c *Client) SetRepeat(ctx context.Context, state, deviceID string) error {
	if state != "track" && state != "context" && state != "off" {
		return errors.New("state must be track, context, or off")
	}
	query := deviceQuery(deviceID)
	if query == nil {
		query = make(url.Values)
	}
	query.Set("state", state)
	return c.Do(ctx, http.MethodPut, "me/player/repeat", query, nil, nil)
}

func (c *Client) SetVolume(ctx context.Context, percent int, deviceID string) error {
	if percent < 0 || percent > 100 {
		return errors.New("volume_percent must be between 0 and 100")
	}
	query := deviceQuery(deviceID)
	if query == nil {
		query = make(url.Values)
	}
	query.Set("volume_percent", strconv.Itoa(percent))
	return c.Do(ctx, http.MethodPut, "me/player/volume", query, nil, nil)
}

func (c *Client) SetShuffle(ctx context.Context, state bool, deviceID string) error {
	query := deviceQuery(deviceID)
	if query == nil {
		query = make(url.Values)
	}
	query.Set("state", strconv.FormatBool(state))
	return c.Do(ctx, http.MethodPut, "me/player/shuffle", query, nil, nil)
}

func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	var result struct {
		Devices []Device `json:"devices"`
	}
	err := c.Do(ctx, http.MethodGet, "me/player/devices", nil, nil, &result)
	return result.Devices, err
}

func (c *Client) TransferPlayback(ctx context.Context, deviceID string, play bool) error {
	if deviceID == "" {
		return errors.New("device_id is required")
	}
	return c.Do(ctx, http.MethodPut, "me/player", nil, map[string]any{"device_ids": []string{deviceID}, "play": play}, nil)
}

func (c *Client) AddToQueue(ctx context.Context, uri, deviceID string) error {
	if uri == "" {
		return errors.New("uri is required")
	}
	query := deviceQuery(deviceID)
	if query == nil {
		query = make(url.Values)
	}
	query.Set("uri", uri)
	return c.Do(ctx, http.MethodPost, "me/player/queue", query, nil, nil)
}

func (c *Client) Queue(ctx context.Context) (Queue, error) {
	var result Queue
	err := c.Do(ctx, http.MethodGet, "me/player/queue", nil, nil, &result)
	return result, err
}
