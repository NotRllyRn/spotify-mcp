package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const SpotifyAuthorizeURL = "https://accounts.spotify.com/authorize"

var scopes = []string{
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
	"user-read-recently-played",
	"user-top-read",
	"playlist-read-private",
	"playlist-read-collaborative",
	"playlist-modify-private",
	"playlist-modify-public",
	"user-library-read",
	"user-library-modify",
}

type AuthorizationConfig struct {
	ClientID    string
	RedirectURI string
	Bind        string
	Port        int
}

type callbackResult struct {
	code string
	err  error
}

func Authorize(ctx context.Context, manager *TokenManager, cfg AuthorizationConfig) error {
	verifier, err := randomURLSafe(64)
	if err != nil {
		return err
	}
	state, err := randomURLSafe(32)
	if err != nil {
		return err
	}
	redirect, err := url.Parse(cfg.RedirectURI)
	if err != nil {
		return fmt.Errorf("parse redirect URI: %w", err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(cfg.Bind, strconv.Itoa(cfg.Port)))
	if err != nil {
		return fmt.Errorf("listen for Spotify callback: %w", err)
	}
	defer listener.Close()

	result := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+redirect.Path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid OAuth state", http.StatusBadRequest)
			return
		}
		if message := r.URL.Query().Get("error"); message != "" {
			result <- callbackResult{err: fmt.Errorf("Spotify authorization failed: %s", message)}
			http.Error(w, "Authorization failed. You may close this window.", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			return
		}
		result <- callbackResult{code: code}
		_, _ = fmt.Fprintln(w, "Spotify authorization complete. You may close this window.")
	})
	callbackServer := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := callbackServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case result <- callbackResult{err: err}:
			default:
			}
		}
	}()

	query := url.Values{
		"client_id":             {cfg.ClientID},
		"response_type":         {"code"},
		"redirect_uri":          {cfg.RedirectURI},
		"scope":                 {strings.Join(scopes, " ")},
		"state":                 {state},
		"code_challenge_method": {"S256"},
		"code_challenge":        {Challenge(verifier)},
	}
	fmt.Printf("Open this URL to authorize Spotify:\n%s?%s\n", SpotifyAuthorizeURL, query.Encode())

	select {
	case <-ctx.Done():
		return ctx.Err()
	case callback := <-result:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = callbackServer.Shutdown(shutdownCtx)
		if callback.err != nil {
			return callback.err
		}
		return manager.ExchangeCode(ctx, callback.code, verifier, cfg.RedirectURI)
	}
}
