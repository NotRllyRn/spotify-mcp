package server

import (
	"context"

	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type playInput struct {
	DeviceID       string   `json:"device_id,omitempty"`
	ContextURI     string   `json:"context_uri,omitempty"`
	URIs           []string `json:"uris,omitempty"`
	OffsetPosition int      `json:"offset_position,omitempty"`
	PositionMS     *int     `json:"position_ms,omitempty"`
}
type seekInput struct {
	DeviceID   string `json:"device_id,omitempty"`
	PositionMS int    `json:"position_ms"`
}
type repeatInput struct {
	DeviceID string `json:"device_id,omitempty"`
	State    string `json:"state"`
}
type volumeInput struct {
	DeviceID      string `json:"device_id,omitempty"`
	VolumePercent int    `json:"volume_percent"`
}
type shuffleInput struct {
	DeviceID string `json:"device_id,omitempty"`
	State    bool   `json:"state"`
}
type transferInput struct {
	DeviceID string `json:"device_id"`
	Play     bool   `json:"play,omitempty"`
}
type queueInput struct {
	URI      string `json:"uri"`
	DeviceID string `json:"device_id,omitempty"`
}

func registerPlaybackTools(server *mcp.Server, client *spotify.Client) {
	addTool(server, "get_playback_state", "Get current Spotify playback state.", func(ctx context.Context, _ emptyInput) (spotify.PlaybackState, error) {
		return client.PlaybackState(ctx)
	})
	addTool(server, "play", "Start or resume playback.", func(ctx context.Context, in playInput) (acknowledgment, error) {
		return acknowledgment{true}, client.Play(ctx, in.DeviceID, in.ContextURI, in.URIs, in.OffsetPosition, in.PositionMS)
	})
	addTool(server, "pause", "Pause playback.", func(ctx context.Context, in deviceInput) (acknowledgment, error) {
		return acknowledgment{true}, client.Pause(ctx, in.DeviceID)
	})
	addTool(server, "next_track", "Skip to the next track.", func(ctx context.Context, in deviceInput) (acknowledgment, error) {
		return acknowledgment{true}, client.Next(ctx, in.DeviceID)
	})
	addTool(server, "previous_track", "Return to the previous track.", func(ctx context.Context, in deviceInput) (acknowledgment, error) {
		return acknowledgment{true}, client.Previous(ctx, in.DeviceID)
	})
	addTool(server, "seek", "Seek to a playback position.", func(ctx context.Context, in seekInput) (acknowledgment, error) {
		return acknowledgment{true}, client.Seek(ctx, in.PositionMS, in.DeviceID)
	})
	addTool(server, "set_repeat", "Set repeat mode to track, context, or off.", func(ctx context.Context, in repeatInput) (acknowledgment, error) {
		return acknowledgment{true}, client.SetRepeat(ctx, in.State, in.DeviceID)
	})
	addTool(server, "set_volume", "Set playback volume from 0 to 100.", func(ctx context.Context, in volumeInput) (acknowledgment, error) {
		return acknowledgment{true}, client.SetVolume(ctx, in.VolumePercent, in.DeviceID)
	})
	addTool(server, "set_shuffle", "Set shuffle state.", func(ctx context.Context, in shuffleInput) (acknowledgment, error) {
		return acknowledgment{true}, client.SetShuffle(ctx, in.State, in.DeviceID)
	})
	addTool(server, "get_devices", "List available Spotify Connect devices.", func(ctx context.Context, _ emptyInput) ([]spotify.Device, error) {
		return client.Devices(ctx)
	})
	addTool(server, "transfer_playback", "Transfer playback to a device.", func(ctx context.Context, in transferInput) (acknowledgment, error) {
		return acknowledgment{true}, client.TransferPlayback(ctx, in.DeviceID, in.Play)
	})
	addTool(server, "add_to_queue", "Add a Spotify URI to the queue.", func(ctx context.Context, in queueInput) (acknowledgment, error) {
		return acknowledgment{true}, client.AddToQueue(ctx, in.URI, in.DeviceID)
	})
	addTool(server, "get_queue", "Get the current playback queue.", func(ctx context.Context, _ emptyInput) (spotify.Queue, error) {
		return client.Queue(ctx)
	})
}
