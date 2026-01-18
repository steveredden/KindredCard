/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package models

import "time"

// APIToken represents an API token for programmatic access
type APIToken struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	TokenHash  string     `json:"-"` // Never expose the hash in JSON
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// APITokenWithRaw is used only during token creation to return the raw token once
type APITokenWithRaw struct {
	APIToken
	RawToken string `json:"token"` // Only set during creation
}

// CreateAPITokenRequest represents the request to create a new token
type CreateAPITokenRequest struct {
	Name      string     `json:"name" binding:"required,min=1,max=255"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APITokenListResponse represents a token in list view (safe to show)
type APITokenListResponse struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
	IsExpired  bool       `json:"is_expired"`
	Prefix     string     `json:"prefix"` // First 8 chars for identification
}

type TokenTestResponse struct {
	Authenticated bool   `json:"authenticated" example:"true"`
	UserID        int    `json:"user_id" example:"1"`
	Message       string `json:"message" example:"API token is valid"`
}
