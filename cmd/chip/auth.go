package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	scalekit "github.com/scalekit-inc/scalekit-sdk-go/v2"
)

type oauthMiddleware struct {
	client    scalekit.Scalekit
	config    *AuthConfig
	wwwHeader string
}

func newOAuthMiddleware(config *AuthConfig) *oauthMiddleware {
	client := scalekit.NewScalekitClient(
		config.EnvironmentURL,
		config.ClientID,
		config.ClientSecret,
	)

	metadataURL := strings.TrimRight(config.ResourceURL, "/") + "/.well-known/oauth-protected-resource"
	wwwHeader := fmt.Sprintf(`Bearer realm="OAuth", resource_metadata="%s"`, metadataURL)

	return &oauthMiddleware{
		client:    client,
		config:    config,
		wwwHeader: wwwHeader,
	}
}

func (m *oauthMiddleware) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.well-known") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.unauthorized(w)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		valid, err := m.client.ValidateTokenWithOptions(r.Context(), tokenStr, &scalekit.ValidateTokenOptions{
			Audience: []string{m.config.ResourceURL},
		})
		if err != nil || !valid {
			slog.InfoContext(r.Context(), fmt.Sprintf("Token validation failed: %v", err))
			m.unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *oauthMiddleware) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", m.wwwHeader)
	w.WriteHeader(http.StatusUnauthorized)
}

type resourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
}

func wellKnownHandler(config *AuthConfig) http.HandlerFunc {
	metadata := resourceMetadata{
		Resource:               config.ResourceURL,
		AuthorizationServers:   config.AuthorizationServers,
		BearerMethodsSupported: []string{"header"},
	}
	data, _ := json.Marshal(metadata)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}
