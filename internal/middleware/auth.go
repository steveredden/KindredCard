/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/steveredden/KindredCard/internal/auth"
	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"
const SessionToken string = "session_token"

// GetUserFromContext extracts user from request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {

	}
	return user, ok
}

// GetUserFromContext extracts user from request context
func GetTokenFromCurrentSession(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(SessionToken)
	if err == nil {
		token := cookie.Value[:len(cookie.Value)-65] // Remove ".signature" part (64 hex chars + 1 dot)
		return token, true
	}
	return "", false
}

// AuthMiddleware checks for valid session token (Browser Client)
// Halts on failure by REDIRECTING to /login
func AuthMiddleware(database *db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("session_token")
			if err != nil {
				// No cookie, redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			appKey := os.Getenv("APP_KEY")
			if appKey == "" {
				logger.Error("[MIDDLEWARE] Error APP_KEY not set in AuthMiddleware")
				http.Error(w, "Server misconfiguration", http.StatusInternalServerError)
				return
			}

			// ADJUSTMENT 1: Use a function that verifies the signature AND extracts the user ID
			// Assuming this function returns (token_payload, userID, err)
			tokenPayload, userID, err := auth.VerifyAndExtractUserID(cookie.Value, appKey)
			if err != nil {
				logger.Error("[MIDDLEWARE] Invalid token signature in AuthMiddleware: %v", err)
				// Clear invalid cookie and redirect
				http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// ADJUSTMENT 2: Use the full token payload for database lookup
			// (The tokenPayload now includes the UserID prefix, e.g., "123-randomstring")
			session, err := database.GetSessionByToken(tokenPayload)
			if err != nil || session == nil || session.UserID != userID {
				// Invalid session (not found, expired, or UserID mismatch)
				http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Get user by the ID extracted from the token (faster than going through session)
			user, err := database.GetUserByID(userID)
			if err != nil || user == nil {
				// User not found (might be deleted)
				http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// SUCCESS: Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// -----------------------------------------------------------------------------
// API/HYBRID AUTH MIDDLEWARE
// -----------------------------------------------------------------------------

// APIAuthMiddleware checks for multiple auth methods (API/Bearer/Session Cookie)
// Halts on failure by returning a 401 Unauthorized response.
func APIAuthMiddleware(database *db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var user *models.User
			var authMethodFound bool

			// Get APP_KEY once
			appKey := os.Getenv("APP_KEY")
			if appKey == "" {
				logger.Error("[MIDDLEWARE] Error APP_KEY not set in APIAuthMiddleware")
				http.Error(w, "Server misconfiguration", http.StatusInternalServerError)
				return
			}

			// 1. TRY API TOKEN (via "session" header)
			apiToken := r.Header.Get("session")
			if apiToken != "" {
				authMethodFound = true
				userID, err := database.ValidateAPIToken(apiToken)
				if err == nil && userID > 0 {
					user, _ = database.GetUserByID(userID)
				}
			}

			// 2. TRY SESSION/BEARER TOKEN
			if user == nil { // Only proceed if not already authenticated via API token
				var signedToken string

				// Try cookie first
				cookie, err := r.Cookie("session_token")
				if err == nil {
					signedToken = cookie.Value
				} else {
					// Try Authorization header
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						signedToken = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if signedToken != "" {
					authMethodFound = true
					tokenPayload, userID, err := auth.VerifyAndExtractUserID(signedToken, appKey)

					if err == nil {
						session, err := database.GetSessionByToken(tokenPayload)
						// Verify session exists AND the UserID extracted from the token matches the session UserID
						if err == nil && session != nil && session.UserID == userID {
							user, _ = database.GetUserByID(userID)
						}
					}
				}
			}

			// 3. HALTING LOGIC
			if user == nil {
				// If any auth method was attempted, return 401, otherwise return the appropriate web redirect
				if authMethodFound || strings.HasPrefix(r.URL.Path, "/api/") {
					apiUnauthorized(w) // HALT with 401 JSON
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther) // HALT with redirect
				}
				return
			}

			// SUCCESS: Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CardDAVAuthMiddleware handles HTTP Basic Auth for CardDAV
func CardDAVAuthMiddleware(database *db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Basic Auth credentials
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Set("WWW-Authenticate", `Basic realm="KindredCard CardDAV"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate credentials
			user, err := database.ValidateUserCredentials(username, password)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="KindredCard CardDAV"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetupCheckMiddleware redirects to setup if not complete
func SetupCheckMiddleware(database *db.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip setup check for setup and static routes
			if strings.HasPrefix(r.URL.Path, "/setup") ||
				strings.HasPrefix(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if setup is complete
			isComplete, err := database.IsSetupComplete()
			if err != nil || !isComplete {
				logger.Error("[MIDDLEWARE] No users have been created; redirecting to first time /setup.html: %v", err)
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// apiUnauthorized sends a 401 Unauthorized JSON response.
func apiUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": "Unauthorized: Invalid session or API token."}`))
}
