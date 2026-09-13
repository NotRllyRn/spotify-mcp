package server

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type emptyInput struct{}
type deviceInput struct {
	DeviceID string `json:"device_id,omitempty"`
}
type pageInput struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}
type idInput struct {
	ID string `json:"id"`
}
type idPageInput struct {
	ID     string `json:"id"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
type acknowledgment struct {
	Success bool `json:"success"`
}

type toolFailure struct{ err error }

func (e toolFailure) Error() string {
	output := struct {
		Error             string `json:"error"`
		Message           string `json:"message"`
		RetryAfterSeconds int64  `json:"retry_after_seconds,omitempty"`
	}{Error: "invalid_request", Message: e.err.Error()}
	var apiError *spotify.APIError
	if errors.As(e.err, &apiError) {
		output.Error = apiError.Code
		output.Message = apiError.Message
		output.RetryAfterSeconds = int64(apiError.RetryAfter / time.Second)
	}
	data, _ := json.Marshal(output)
	return string(data)
}

func addTool[In, Out any](server *mcp.Server, name, description string, handler func(context.Context, In) (Out, error)) {
	mcp.AddTool(server, &mcp.Tool{Name: name, Description: description}, func(ctx context.Context, _ *mcp.CallToolRequest, input In) (*mcp.CallToolResult, Out, error) {
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		output, err := handler(ctx, input)
		if err != nil {
			return nil, output, toolFailure{err}
		}
		return nil, output, nil
	})
}

func normalizedPage(input pageInput, defaultLimit, max int) (int, int, error) {
	if input.Limit == 0 {
		input.Limit = defaultLimit
	}
	if input.Limit < 1 || input.Limit > max {
		return 0, 0, errors.New("limit is out of range")
	}
	if input.Offset < 0 {
		return 0, 0, errors.New("offset must be at least 0")
	}
	return input.Limit, input.Offset, nil
}

func RegisterTools(server *mcp.Server, client *spotify.Client) {
	registerPlaybackTools(server, client)
	registerSearchTools(server, client)
	registerPlaylistTools(server, client)
	registerLibraryTools(server, client)
}
