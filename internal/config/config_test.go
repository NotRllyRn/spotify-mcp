package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SPOTIFY_CALLBACK_PORT", "")
	t.Setenv("SPOTIFY_REDIRECT_URI", "")
	t.Setenv("LOG_LEVEL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SpotifyCallbackPort != 8888 || cfg.SpotifyRedirectURI != defaultRedirectURI {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct{ name, value string }{
		{"SPOTIFY_CALLBACK_PORT", "70000"},
		{"SPOTIFY_REDIRECT_URI", "/callback"},
		{"LOG_LEVEL", "verbose"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPOTIFY_CALLBACK_PORT", "")
			t.Setenv("SPOTIFY_REDIRECT_URI", "")
			t.Setenv("LOG_LEVEL", "")
			t.Setenv(tt.name, tt.value)
			if _, err := Load(); err == nil {
				t.Fatal("Load() succeeded, want error")
			}
		})
	}
}

func TestValidation(t *testing.T) {
	if err := (Config{}).ValidateAuth(); err == nil {
		t.Fatal("ValidateAuth() succeeded without client ID")
	}
	if err := (Config{SpotifyClientID: "client"}).ValidateServe(); err == nil {
		t.Fatal("ValidateServe() succeeded without MCP token")
	}
	if err := (Config{SpotifyClientID: "client", MCPAuthToken: "token"}).ValidateServe(); err != nil {
		t.Fatal(err)
	}
}
