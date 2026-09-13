# spotify-mcp

A small self-hosted MCP server for controlling Spotify.

## Setup

1. Create a Spotify Web API app and add:
   `http://127.0.0.1:8888/callback`
2. Copy and edit the environment file:
   `cp example.env .env`
3. Authorize Spotify:
   `docker compose --env-file .env -f compose.example.yaml run --rm --service-ports spotify-mcp-auth`
4. Start:
   `docker compose --env-file .env -f compose.example.yaml up -d`

MCP endpoint:

`http://127.0.0.1:8765/mcp`

Use `Authorization: Bearer <MCP_AUTH_TOKEN>`.

See [DOCS.md](DOCS.md) for configuration, tools, reverse-proxy notes, and troubleshooting.
