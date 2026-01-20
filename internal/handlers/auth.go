/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package handlers

import (
	"net/http"
	"os"

	"github.com/steveredden/KindredCard/internal/auth"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/session"
)

// Authentication Handlers

func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, r, "login.html", map[string]interface{}{
		"Title": "Login",
	})
}

func (h *Handler) ProcessLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.db.ValidateUserCredentials(email, password)
	if err != nil {
		h.renderTemplate(w, r, "login.html", map[string]interface{}{
			"Title": "Login",
			"Error": "Invalid email or password",
		})
		return
	}

	// Get APP_KEY for token signing
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		logger.Error("[HANDLER] Missing APP_KEY environment variable")
		http.Error(w, "Server misconfiguration - APP_KEY not set", http.StatusInternalServerError)
		return
	}

	// Create signed session token
	signedToken, err := auth.GenerateSignedToken(appKey, user.ID)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Extract the unsigned token for database storage
	token := signedToken[:len(signedToken)-65] // Remove ".signature" part (64 hex chars + 1 dot)

	sessionInfo := session.GetSessionInfo(r)
	expiresAt := auth.GetSessionExpiry()

	dbSession := &models.Session{
		UserID:       user.ID,
		Token:        token,
		UserAgent:    sessionInfo.UserAgent,
		Browser:      sessionInfo.Browser,
		BrowserVer:   sessionInfo.BrowserVer,
		OS:           sessionInfo.OS,
		Device:       sessionInfo.Device,
		IsMobile:     sessionInfo.IsMobile,
		IPAddress:    sessionInfo.IPAddress,
		Referer:      sessionInfo.Referer,
		Language:     sessionInfo.Language,
		LoginTime:    sessionInfo.LoginTime,
		LastActivity: sessionInfo.LoginTime,
		ExpiresAt:    expiresAt,
	}

	err = h.db.CreateSession(dbSession)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    signedToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure:   false, // Set true in production with HTTPS
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	token, ok := middleware.GetTokenFromCurrentSession(r)
	if !ok {
		http.Error(w, "Failed to delete session Token", http.StatusInternalServerError)
		return
	}

	err := h.db.RevokeSession(user.ID, token)
	if err != nil {
		http.Error(w, "Failed to delete session Token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Setup Handlers

func (h *Handler) ShowSetup(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, r, "setup.html", map[string]interface{}{
		"Title": "Setup",
		"Error": "",
	})
}

func (h *Handler) ProcessSetup(w http.ResponseWriter, r *http.Request) {

	// Get APP_KEY for token signing
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		http.Error(w, "Server misconfiguration - APP_KEY not set", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	if password != confirm {
		h.renderTemplate(w, r, "setup.html", map[string]interface{}{
			"Title": "Setup",
			"Error": "Passwords don't match",
		})
		return
	}

	if len(password) < 8 {
		h.renderTemplate(w, r, "setup.html", map[string]interface{}{
			"Title": "Setup",
			"Error": "Password must be at least 8 characters",
		})
		return
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.db.CreateUser(email, hash)
	if err != nil {
		h.renderTemplate(w, r, "setup.html", map[string]interface{}{
			"Title": "Setup",
			"Error": "Email already exists or database error",
		})
		return
	}

	// Mark setup complete
	h.db.MarkSetupComplete(user.ID)

	// Auto-login with signed token
	signedToken, err := auth.GenerateSignedToken(appKey, user.ID)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Extract the unsigned token for database storage
	token := signedToken[:len(signedToken)-65] // Remove ".signature" part (64 hex chars + 1 dot)

	sessionInfo := session.GetSessionInfo(r)
	expiresAt := auth.GetSessionExpiry()

	dbSession := &models.Session{
		UserID:       user.ID,
		Token:        token,
		UserAgent:    sessionInfo.UserAgent,
		Browser:      sessionInfo.Browser,
		BrowserVer:   sessionInfo.BrowserVer,
		OS:           sessionInfo.OS,
		Device:       sessionInfo.Device,
		IsMobile:     sessionInfo.IsMobile,
		IPAddress:    sessionInfo.IPAddress,
		Referer:      sessionInfo.Referer,
		Language:     sessionInfo.Language,
		LoginTime:    sessionInfo.LoginTime,
		LastActivity: sessionInfo.LoginTime,
		ExpiresAt:    expiresAt,
	}

	err = h.db.CreateSession(dbSession)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    signedToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // Recommended in production
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RotatePassword handles a user's request to change their password.
func (h *Handler) RotatePassword(w http.ResponseWriter, r *http.Request) {

	// 1. **Authentication Check & Get User**
	// Assumes this handler is protected by authentication middleware
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		// Should be caught by middleware, but good safety check for API routes
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. **Extract Form Values**
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmNewPassword := r.FormValue("confirm_password")

	// 3. **Validation Checks**
	if newPassword != confirmNewPassword {
		h.renderTemplate(w, r, "settings.html", map[string]interface{}{
			"Error": "New passwords do not match.",
		})
		return
	}

	if len(newPassword) < 8 {
		h.renderTemplate(w, r, "settings.html", map[string]interface{}{
			"Error": "New password must be at least 8 characters.",
		})
		return
	}

	// 4. **Verify Current Password (Critical Step)**
	// You need the stored hash from the user's database record (user.PasswordHash).
	if err := auth.CheckPassword(currentPassword, user.PasswordHash); err != nil {
		h.renderTemplate(w, r, "settings.html", map[string]interface{}{
			"Error": "Current password is incorrect.",
		})
		return
	}

	// 5. **Hash New Password**
	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		http.Error(w, "Error processing new password", http.StatusInternalServerError)
		return
	}

	// 6. **Update Password in Database**
	err = h.db.UpdatePasswordHash(user.ID, newHash)
	if err != nil {
		logger.Error("DB error updating password for user %d: %v", user.ID, err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	// 7. **SECURITY STEP: Invalidate All Existing Sessions**
	// This immediately forces all old tokens/sessions for this user to become invalid.
	err = h.db.InvalidateAllSessions(user.ID)
	if err != nil {
		logger.Error("DB error invalidating sessions for user %d: %v", user.ID, err)
		// Log the error but continue, as the password change itself succeeded
	}

	// 8. **Generate New Session Token (Auto-login)**
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		http.Error(w, "Server misconfiguration - APP_KEY not set", http.StatusInternalServerError)
		return
	}

	// Generate a new token and create a new session
	signedToken, err := auth.GenerateSignedToken(appKey, user.ID) // Assume GenerateSignedToken takes user.ID now
	if err != nil {
		http.Error(w, "Failed to generate new session token", http.StatusInternalServerError)
		return
	}

	token := signedToken[:len(signedToken)-65] // Remove ".signature" part (64 hex chars + 1 dot)

	sessionInfo := session.GetSessionInfo(r)
	expiresAt := auth.GetSessionExpiry()

	dbSession := &models.Session{
		UserID:       user.ID,
		Token:        token,
		UserAgent:    sessionInfo.UserAgent,
		Browser:      sessionInfo.Browser,
		BrowserVer:   sessionInfo.BrowserVer,
		OS:           sessionInfo.OS,
		Device:       sessionInfo.Device,
		IsMobile:     sessionInfo.IsMobile,
		IPAddress:    sessionInfo.IPAddress,
		Referer:      sessionInfo.Referer,
		Language:     sessionInfo.Language,
		LoginTime:    sessionInfo.LoginTime,
		LastActivity: sessionInfo.LoginTime,
		ExpiresAt:    expiresAt,
	}

	err = h.db.CreateSession(dbSession)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// 9. **Set New Cookie & Redirect**
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    signedToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure: true, // Recommended in production
	})

	// Redirect to the dashboard or settings page with a success message
	http.Redirect(w, r, "/settings#security", http.StatusSeeOther)
}
