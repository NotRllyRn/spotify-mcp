package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const APIBaseURL = "https://api.spotify.com/v1/"

type TokenProvider interface {
	Token(context.Context, bool) (string, error)
}

type Client struct {
	http *http.Client
	auth TokenProvider
	base *url.URL
}

func NewClient(httpClient *http.Client, auth TokenProvider, baseURL string) (*Client, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Spotify base URL: %w", err)
	}
	return &Client{http: httpClient, auth: auth, base: base}, nil
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode Spotify request: %w", err)
		}
	}
	ref := &url.URL{Path: strings.TrimPrefix(path, "/"), RawQuery: query.Encode()}
	endpoint := c.base.ResolveReference(ref).String()
	refreshed, rateRetried := false, false
	for {
		token, err := c.auth.Token(ctx, refreshed)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create Spotify request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		response, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("Spotify request: %w", err)
		}
		data, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read Spotify response: %w", readErr)
		}
		if response.StatusCode == http.StatusUnauthorized && !refreshed {
			refreshed = true
			continue
		}
		if response.StatusCode == http.StatusTooManyRequests {
			delay := retryAfter(response.Header.Get("Retry-After"), time.Now())
			if !rateRetried && delay <= 30*time.Second {
				rateRetried = true
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
					continue
				}
			}
			return &APIError{StatusCode: response.StatusCode, Code: "rate_limited", Message: fmt.Sprintf("Spotify requested a %s retry delay", delay), RetryAfter: delay}
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return decodeAPIError(response.StatusCode, data)
		}
		if out == nil || len(data) == 0 {
			return nil
		}
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode Spotify response: %w", err)
		}
		return nil
	}
}

func retryAfter(value string, now time.Time) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil && date.After(now) {
		return date.Sub(now)
	}
	return time.Second
}

func decodeAPIError(status int, data []byte) error {
	var response struct {
		Error struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
	}
	message := http.StatusText(status)
	if json.Unmarshal(data, &response) == nil && response.Error.Message != "" {
		message = response.Error.Message
	}
	code := "spotify_error"
	switch status {
	case http.StatusForbidden:
		code = "forbidden"
	case http.StatusNotFound:
		code = "not_found"
	case http.StatusUnauthorized:
		code = "unauthorized"
	}
	return &APIError{StatusCode: status, Code: code, Message: message}
}

func requireID(name, value string) error {
	if strings.TrimSpace(value) == "" || strings.Contains(value, "/") {
		return errors.New(name + " is required and must not contain a slash")
	}
	return nil
}

func pagination(limit, offset, max int) (url.Values, error) {
	if limit < 1 || limit > max {
		return nil, fmt.Errorf("limit must be between 1 and %d", max)
	}
	if offset < 0 {
		return nil, errors.New("offset must be at least 0")
	}
	return url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(offset)}}, nil
}
