# spotify-mcp documentation

This Go rewrite is derived from `llyfn/spotify-mcp` and keeps its MIT license and project history. It exposes a stateless Streamable HTTP MCP endpoint and calls Spotify directly.

## Spotify developer-app setup

Create an app in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard), enable the Web API, and register this redirect URI exactly:

```text
http://127.0.0.1:8888/callback
```

Only the app's client ID is needed. No client secret is used. Playback control requires Spotify Premium.

## Docker setup

```bash
cp example.env .env
```

Set `SPOTIFY_CLIENT_ID` and replace `MCP_AUTH_TOKEN` with a long random value. Build and start with:

```bash
docker compose --env-file .env -f compose.example.yaml up -d
```

The default endpoint is `http://127.0.0.1:8765/mcp`. The container runs as UID 65532 with a read-only root filesystem, no Linux capabilities, and a named volume for `/data/token.json`.

## First authorization

Run this once before using Spotify tools:

```bash
docker compose --env-file .env -f compose.example.yaml run --rm --service-ports spotify-mcp auth
```

Open the printed URL and approve access. Spotify redirects the browser to the loopback-only callback port. The command stores only the refresh token and original authorization time in the named volume.

## MCP client configuration

Configure a Streamable HTTP server with:

```json
{
  "url": "http://127.0.0.1:8765/mcp",
  "headers": {
    "Authorization": "Bearer YOUR_MCP_AUTH_TOKEN"
  }
}
```

The exact surrounding configuration depends on the MCP client. Every request needs the bearer header.

## Reverse proxy

Keep `MCP_PUBLISH_ADDR=127.0.0.1` and proxy HTTPS to `http://127.0.0.1:8765`. Forward request bodies and the `Authorization` header without modification. For browser-originated clients, list the exact HTTPS origin in `MCP_ALLOWED_ORIGINS`.

TLS and public access controls belong in Caddy, Nginx, Traefik, or another ingress. To deliberately publish directly on a LAN, set `MCP_PUBLISH_ADDR=0.0.0.0`; the bearer token remains mandatory.

## Environment variables

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `SPOTIFY_CLIENT_ID` | Yes | — | Spotify application client ID |
| `SPOTIFY_REDIRECT_URI` | No | `http://127.0.0.1:8888/callback` | Registered browser redirect URI |
| `SPOTIFY_CALLBACK_PORT` | No | `8888` | Callback host-publish port |
| `SPOTIFY_CALLBACK_BIND` | No | `127.0.0.1` | Internal callback listener bind address |
| `MCP_AUTH_TOKEN` | Serve | — | Bearer token protecting `/mcp` |
| `MCP_LISTEN_ADDR` | No | `127.0.0.1:8080` | Internal HTTP listen address |
| `MCP_PUBLISH_ADDR` | Compose | `127.0.0.1` | Docker host publish address |
| `MCP_PUBLISH_PORT` | Compose | `8765` | Docker host publish port |
| `MCP_ALLOWED_ORIGINS` | No | Empty | Comma-separated exact Origin allowlist |
| `TOKEN_PATH` | No | `/data/token.json` | Refresh-token state path |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, or `error` |

`MCP_ALLOWED_ORIGINS` does not support `*`. Requests without an `Origin` header are accepted after bearer authentication.

## Tool catalog

### Search and metadata

- `search_tracks`, `search_albums`, `search_artists`, `search_playlists`
- `get_track`, `get_album`, `get_artist`, `get_artist_albums`
- `get_playlist`, `get_playlist_items`, `get_my_profile`

Search defaults to 5 results and permits at most 10.

### Personalization

- `get_recently_played`
- `get_top_tracks`, `get_top_artists`

### Playlists

- `list_my_playlists`, `create_playlist`, `update_playlist`
- `add_playlist_items`, `remove_playlist_items`, `reorder_playlist_items`

Playlist mutations are performed sequentially in chunks of 100.

### Library

- `get_saved_tracks`, `get_saved_albums`
- `save_to_library`, `remove_from_library`, `check_library`

Generic library operations accept Spotify URIs and are chunked at 40.

### Playback

- `get_playback_state`, `play`, `pause`
- `next_track`, `previous_track`, `seek`
- `set_repeat`, `set_volume`, `set_shuffle`
- `get_devices`, `transfer_playback`
- `add_to_queue`, `get_queue`

`play` accepts either `context_uri` or `uris`, never both.

## OAuth and token lifecycle

Authorization uses Authorization Code with PKCE. The short-lived access token exists only in memory and is refreshed on demand under a mutex. A Spotify `401` forces one refresh and one retry. Rotated refresh tokens are saved atomically with mode `0600` where supported.

Spotify refresh tokens expire six months after the original authorization. Refreshing an access token does not extend that date.

## Reauthorization

If a tool reports `reauthorization_required`, run the first-authorization command again. The server does not retry `invalid_grant` responses or poll in the background.

## Development Mode limitations

This server targets Spotify's restricted Development Mode API:

- search is limited to 10 results;
- removed bulk metadata endpoints are not exposed;
- playlist contents may be unavailable for playlists the user does not own or collaborate on;
- playlist item APIs use `/items` and `item` fields;
- removed popularity, follower, private-profile, podcast, and audiobook fields are not modeled;
- recommendations, audio features, shows, episodes, audiobooks, and follow mutations are outside this server's scope.

## Troubleshooting

- **401 from `/mcp`:** verify the exact `Bearer` token.
- **403 from `/mcp`:** add the request's exact `Origin` to `MCP_ALLOWED_ORIGINS`.
- **Callback does not open:** verify port 8888 is free and the redirect URI exactly matches the Spotify app.
- **`reauthorization_required`:** rerun `spotify-mcp auth`.
- **Spotify 403:** check app access, granted scopes, playlist ownership, and Premium status for playback.
- **No active device:** open Spotify on a device or call `transfer_playback` with a listed device ID.
- **Rate limited:** retry after the returned delay. Delays above 30 seconds are never slept automatically.

## Local Go development

Go 1.27 or newer is required.

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/spotify-mcp
```

For local authorization, create the token directory and override its path:

```bash
mkdir -p .data
SPOTIFY_CLIENT_ID=... TOKEN_PATH=.data/token.json go run ./cmd/spotify-mcp auth
```

Start locally with `MCP_AUTH_TOKEN`, `SPOTIFY_CLIENT_ID`, and `TOKEN_PATH` set. The server performs no idle polling, scheduled refresh, or background Spotify traffic.
