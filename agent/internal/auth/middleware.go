// Package auth provides HTTP authentication middleware for FORGE.
//
// v1: no-op stub — single-user local installs need no auth.
// If FORGE_SECRET_KEY env var is set, the X-Forge-Key header is validated.
//
// v2+: replace NoOpAuthenticator with a Clerk/SSO implementation that
// satisfies the Authenticator interface. The router wiring in api/router.go
// requires no changes.
package auth

import (
	"encoding/json"
	"net/http"
)

// Authenticator is the interface for auth providers.
// Implement this to swap in Clerk, OIDC, or any other auth backend.
type Authenticator interface {
	Middleware() func(http.Handler) http.Handler
}

// NoOpAuthenticator is the v1 implementation.
// Pass-through if secretKey is empty; otherwise validates X-Forge-Key header.
type NoOpAuthenticator struct {
	secretKey string
}

// New creates the v1 auth middleware.
// secretKey should come from the FORGE_SECRET_KEY env var (empty = no auth).
func New(secretKey string) *NoOpAuthenticator {
	return &NoOpAuthenticator{secretKey: secretKey}
}

// Middleware returns an http middleware that enforces the optional secret key.
// The /api/health endpoint is always allowed regardless of key configuration.
func (a *NoOpAuthenticator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// No secret configured — pure pass-through.
		if a.secretKey == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health check is always public.
			if r.URL.Path == "/api/health" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("X-Forge-Key") != a.secretKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
