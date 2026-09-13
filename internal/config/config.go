package config

import (
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const defaultRedirectURI = "http://127.0.0.1:8888/callback"

type Config struct {
	SpotifyClientID     string
	SpotifyRedirectURI  string
	SpotifyCallbackBind string
	SpotifyCallbackPort int
	TokenPath           string
	MCPListenAddr       string
	MCPAuthToken        string
	MCPAllowedOrigins   map[string]struct{}
	LogLevel            slog.Level
}

func Load() (Config, error) {
	callbackPort, err := intEnv("SPOTIFY_CALLBACK_PORT", 8888)
	if err != nil {
		return Config{}, err
	}
	redirectURI := env("SPOTIFY_REDIRECT_URI", defaultRedirectURI)
	if _, err := url.ParseRequestURI(redirectURI); err != nil {
		return Config{}, errors.New("SPOTIFY_REDIRECT_URI must be a valid URI")
	}
	level := new(slog.LevelVar)
	if err := level.UnmarshalText([]byte(env("LOG_LEVEL", "info"))); err != nil {
		return Config{}, errors.New("LOG_LEVEL must be debug, info, warn, or error")
	}
	return Config{
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyRedirectURI:  redirectURI,
		SpotifyCallbackBind: env("SPOTIFY_CALLBACK_BIND", "127.0.0.1"),
		SpotifyCallbackPort: callbackPort,
		TokenPath:           env("TOKEN_PATH", "/data/token.json"),
		MCPListenAddr:       env("MCP_LISTEN_ADDR", "127.0.0.1:8080"),
		MCPAuthToken:        os.Getenv("MCP_AUTH_TOKEN"),
		MCPAllowedOrigins:   origins(os.Getenv("MCP_ALLOWED_ORIGINS")),
		LogLevel:            level.Level(),
	}, nil
}

func (c Config) ValidateAuth() error {
	if c.SpotifyClientID == "" {
		return errors.New("SPOTIFY_CLIENT_ID is required")
	}
	return nil
}

func (c Config) ValidateServe() error {
	if err := c.ValidateAuth(); err != nil {
		return err
	}
	if c.MCPAuthToken == "" {
		return errors.New("MCP_AUTH_TOKEN is required")
	}
	return nil
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 65535 {
		return 0, errors.New(name + " must be a valid port")
	}
	return n, nil
}

func origins(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for origin := range strings.SplitSeq(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" && origin != "*" {
			result[origin] = struct{}{}
		}
	}
	return result
}
