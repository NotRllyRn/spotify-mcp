package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NotRllyRn/spotify-mcp/internal/auth"
	"github.com/NotRllyRn/spotify-mcp/internal/config"
	"github.com/NotRllyRn/spotify-mcp/internal/server"
	"github.com/NotRllyRn/spotify-mcp/internal/spotify"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command == "version" {
		fmt.Println(version)
		return nil
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel})))
	switch command {
	case "serve":
		return serve(cfg)
	case "auth":
		return authorize(cfg)
	default:
		return fmt.Errorf("unknown command %q (use serve, auth, or version)", command)
	}
}

func sharedHTTPClient() *http.Client {
	return &http.Client{Transport: &http.Transport{MaxIdleConns: 20, MaxIdleConnsPerHost: 10, IdleConnTimeout: 90 * time.Second}}
}

func authorize(cfg config.Config) error {
	if err := cfg.ValidateAuth(); err != nil {
		return err
	}
	store := auth.Store{Path: cfg.TokenPath}
	if err := store.CheckWritable(); err != nil {
		return err
	}
	manager := auth.NewTokenManager(sharedHTTPClient(), store, cfg.SpotifyClientID, auth.SpotifyTokenURL)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return auth.Authorize(ctx, manager, auth.AuthorizationConfig{
		ClientID: cfg.SpotifyClientID, RedirectURI: cfg.SpotifyRedirectURI,
		Bind: cfg.SpotifyCallbackBind, Port: cfg.SpotifyCallbackPort,
	})
}

func serve(cfg config.Config) error {
	if err := cfg.ValidateServe(); err != nil {
		return err
	}
	if err := (auth.Store{Path: cfg.TokenPath}).CheckWritable(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := sharedHTTPClient()
	manager := auth.NewTokenManager(httpClient, auth.Store{Path: cfg.TokenPath}, cfg.SpotifyClientID, auth.SpotifyTokenURL)
	spotifyClient, err := spotify.NewClient(httpClient, manager, spotify.APIBaseURL)
	if err != nil {
		return err
	}
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "spotify-mcp", Version: version}, nil)
	server.RegisterTools(mcpServer, spotifyClient)
	httpServer := server.NewHTTP(cfg.MCPListenAddr, cfg.MCPAuthToken, cfg.MCPAllowedOrigins, mcpServer)
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "address", cfg.MCPListenAddr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
