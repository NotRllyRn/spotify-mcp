package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const SpotifyTokenURL = "https://accounts.spotify.com/api/token"

var ErrReauthorizationRequired = errors.New("reauthorization_required: run spotify-mcp auth")

type TokenManager struct {
	http      *http.Client
	store     Store
	clientID  string
	tokenURL  string
	mu        sync.Mutex
	access    string
	expiresAt time.Time
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	Description  string `json:"error_description"`
}

func NewTokenManager(client *http.Client, store Store, clientID, tokenURL string) *TokenManager {
	return &TokenManager{http: client, store: store, clientID: clientID, tokenURL: tokenURL}
}

func (m *TokenManager) Token(ctx context.Context, force bool) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !force && m.access != "" && time.Until(m.expiresAt) > time.Minute {
		return m.access, nil
	}
	state, err := m.store.Load()
	if err != nil {
		return "", err
	}
	if time.Now().After(state.AuthorizedAt.AddDate(0, 6, 0)) {
		return "", ErrReauthorizationRequired
	}
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {state.RefreshToken},
		"client_id":     {m.clientID},
	}
	token, err := m.exchange(ctx, values)
	if err != nil {
		return "", err
	}
	if token.RefreshToken != "" && token.RefreshToken != state.RefreshToken {
		state.RefreshToken = token.RefreshToken
		if err := m.store.Save(state); err != nil {
			return "", err
		}
	}
	m.access = token.AccessToken
	m.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return m.access, nil
}

func (m *TokenManager) ExchangeCode(ctx context.Context, code, verifier, redirectURI string) error {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {m.clientID},
		"code_verifier": {verifier},
	}
	token, err := m.exchange(ctx, values)
	if err != nil {
		return err
	}
	if token.RefreshToken == "" {
		return errors.New("Spotify did not return a refresh token")
	}
	return m.store.Save(State{RefreshToken: token.RefreshToken, AuthorizedAt: time.Now().UTC()})
}

func (m *TokenManager) exchange(ctx context.Context, values url.Values) (tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := m.http.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("request token: %w", err)
	}
	defer response.Body.Close()
	var token tokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&token); err != nil {
		return tokenResponse{}, fmt.Errorf("decode token response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if token.Error == "invalid_grant" {
			return tokenResponse{}, ErrReauthorizationRequired
		}
		if token.Description == "" {
			token.Description = http.StatusText(response.StatusCode)
		}
		return tokenResponse{}, fmt.Errorf("Spotify token request failed: %s", token.Description)
	}
	if token.AccessToken == "" || token.ExpiresIn < 1 {
		return tokenResponse{}, errors.New("Spotify returned an invalid access token")
	}
	return token, nil
}
