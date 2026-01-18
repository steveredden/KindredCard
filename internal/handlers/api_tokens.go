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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// CreateAPIToken godoc
//
//	@Summary		Create API token
//	@Description	Generate a new API token for programmatic access
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			token	body		models.CreateAPITokenRequest	true	"Token details"
//	@Success		201		{object}	models.APIToken					"Created token (full token only shown once)"
//	@Failure		400		{object}	map[string]string				"Invalid request body"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Failure		500		{object}	map[string]string				"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/tokens [post]
func (h *Handler) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateAPITokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("[HANDLER] Unable to decode json: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate name
	if req.Name == "" || len(req.Name) > 255 {
		logger.Error("[HANDLER] Token name is required and must be less than 255 characters")
		http.Error(w, "Token name is required and must be less than 255 characters", http.StatusBadRequest)
		return
	}

	// Create the token
	tokenWithRaw, err := h.db.CreateAPIToken(user.ID, req.Name, req.ExpiresAt)
	if err != nil {
		http.Error(w, "Failed to create API token", http.StatusInternalServerError)
		return
	}

	// Return the token (ONLY time the raw token is exposed)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tokenWithRaw)
}

// ListAPITokens godoc
//
//	@Summary		List API tokens
//	@Description	Get all API tokens for the authenticated user
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		models.APIToken		"List of API tokens (tokens are masked)"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/tokens [get]
func (h *Handler) ListAPITokens(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokens, err := h.db.GetAPITokensByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch API tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

// GetAPIToken gets a single API token by ID
// GET /api/v1/tokens/:id
func (h *Handler) GetAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tokenIDStr := vars["id"]

	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	token, err := h.db.GetAPITokenByID(user.ID, tokenID)
	if err != nil {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// RevokeAPIToken revokes (deactivates) an API token
// POST /api/v1/tokens/:id/revoke
func (h *Handler) RevokeAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tokenIDStr := vars["id"]

	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	err = h.db.RevokeAPIToken(user.ID, tokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Token revoked successfully",
	})
}

// DeleteAPIToken godoc
//
//	@Summary		Revoke API token
//	@Description	Revoke an API token to prevent further use
//	@Tags			tokens
//	@Accept			json
//	@Produce		json
//	@Param			tokenId	path	int	true	"Token ID"	minimum(1)
//	@Success		204		"Token revoked successfully"
//	@Failure		400		{object}	map[string]string	"Invalid token ID"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		404		{object}	map[string]string	"Token not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/tokens/{tokenId} [delete]
func (h *Handler) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tokenIDStr := vars["id"]

	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteAPIToken(user.ID, tokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestAPIToken godoc
//
//	@Summary		Validate API Token
//	@Description	Checks if the provided API token is valid and returns the authenticated user's ID.
//	@Tags			tokens
//	@Produce		json
//	@Success		200	{object}	models.TokenTestResponse	"Token is valid"
//	@Failure		401	{object}	map[string]string			"Unauthorized - Invalid or missing token"
//	@Failure		500	{object}	map[string]string			"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/tokens/validate [get]
func (h *Handler) TestAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user_id":       user.ID,
		"message":       "API token is valid",
	})
}
