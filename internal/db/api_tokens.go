/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package db

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

// GenerateAPIToken generates a cryptographically secure random token
// Format: kc_live_<32 random bytes base64url encoded> (total ~50 chars)
func GenerateAPIToken() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode as base64url (URL-safe)
	encoded := base64.RawURLEncoding.EncodeToString(b)

	// Prefix for identification and security
	token := fmt.Sprintf("kc_live_%s", encoded)

	return token, nil
}

// SignToken creates an HMAC signature of the token using APP_KEY
// Returns: signature (base64) to be stored alongside the hash
func SignToken(token string, appKey string) string {
	h := hmac.New(sha256.New, []byte(appKey))
	h.Write([]byte(token))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

// VerifyTokenSignature verifies the HMAC signature of a token
func VerifyTokenSignature(token string, signature string, appKey string) bool {
	expectedSignature := SignToken(token, appKey)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// HashToken creates a SHA-256 hash of the token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// CreateAPIToken creates a new API token for a user
// Returns the token WITH the raw token (only time it's exposed)
func (d *Database) CreateAPIToken(userID int, name string, expiresAt *time.Time) (*models.APITokenWithRaw, error) {
	logger.Debug("[DATABASE] Begin CreateAPIToken(userID:%d, name:%s, expiresAt:%v)", userID, name, expiresAt)

	// Get APP_KEY for HMAC signing
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		return nil, fmt.Errorf("APP_KEY not set")
	}

	// Generate raw token
	rawToken, err := GenerateAPIToken()
	if err != nil {
		return nil, err
	}

	// Hash the token for storage
	tokenHash := HashToken(rawToken)

	// Sign the token with HMAC
	signature := SignToken(rawToken, appKey)

	// Insert into database
	query := `
		INSERT INTO api_tokens (user_id, token_hash, token_signature, name, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, name, last_used_at, created_at, expires_at, is_active
	`

	var token models.APIToken
	var lastUsedAt sql.NullTime
	var expiresAtDB sql.NullTime

	err = d.db.QueryRow(query, userID, tokenHash, signature, name, expiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.Name,
		&lastUsedAt,
		&token.CreatedAt,
		&expiresAtDB,
		&token.IsActive,
	)
	if err != nil {
		logger.Error("[DATABASE] Error inserting api token: %v", err)
		return nil, fmt.Errorf("failed to create API token: %w", err)
	}

	// Convert nullable fields
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAtDB.Valid {
		token.ExpiresAt = &expiresAtDB.Time
	}

	// Return with raw token (only time it's available)
	return &models.APITokenWithRaw{
		APIToken: token,
		RawToken: rawToken,
	}, nil
}

// ValidateAPIToken checks if a token is valid and returns the associated user ID
// Also updates the last_used_at timestamp
// Now includes HMAC signature verification for extra security
func (d *Database) ValidateAPIToken(rawToken string) (int, error) {
	logger.Debug("[DATABASE] Begin ValidateAPIToken(rawToken:--)")

	// Get APP_KEY for HMAC verification
	appKey := os.Getenv("APP_KEY")
	if appKey == "" {
		return 0, fmt.Errorf("APP_KEY not set")
	}

	// Validate token format
	if !strings.HasPrefix(rawToken, "kc_live_") {
		return 0, fmt.Errorf("invalid token format")
	}

	tokenHash := HashToken(rawToken)

	// Query to get token details including signature
	query := `
		SELECT user_id, token_signature
		FROM api_tokens
		WHERE token_hash = $1
			AND is_active = true
			AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`

	var userID int
	var storedSignature string
	err := d.db.QueryRow(query, tokenHash).Scan(&userID, &storedSignature)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		logger.Error("[DATABASE] Error selecting api token: %v", err)
		return 0, fmt.Errorf("failed to validate token: %w", err)
	}

	// Verify HMAC signature
	if !VerifyTokenSignature(rawToken, storedSignature, appKey) {
		return 0, fmt.Errorf("invalid token signature")
	}

	// Update last_used_at
	updateQuery := `
		UPDATE api_tokens
		SET last_used_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1
	`
	_, err = d.db.Exec(updateQuery, tokenHash)
	if err != nil {
		// Log but don't fail - usage tracking is non-critical
		logger.Error("[DATABASE] Error updating api token: %v", err)
	}

	return userID, nil
}

// GetAPITokensByUserID retrieves all API tokens for a user
func (d *Database) GetAPITokensByUserID(userID int) ([]models.APITokenListResponse, error) {
	logger.Debug("[DATABASE] Begin GetAPITokensByUserID(userID:%d)", userID)

	query := `
		SELECT 
			id, 
			name, 
			token_hash,
			last_used_at, 
			created_at, 
			expires_at, 
			is_active,
			CASE 
				WHEN expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP THEN true
				ELSE false
			END as is_expired
		FROM api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting api token: %v", err)
		return nil, fmt.Errorf("failed to fetch API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []models.APITokenListResponse
	for rows.Next() {
		var token models.APITokenListResponse
		var lastUsedAt sql.NullTime
		var expiresAt sql.NullTime
		var tokenHash string

		err := rows.Scan(
			&token.ID,
			&token.Name,
			&tokenHash,
			&lastUsedAt,
			&token.CreatedAt,
			&expiresAt,
			&token.IsActive,
			&token.IsExpired,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning api token: %v", err)
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}

		// Convert nullable fields
		if lastUsedAt.Valid {
			token.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			token.ExpiresAt = &expiresAt.Time
		}

		// Generate prefix from hash for display (first 8 chars)
		if len(tokenHash) >= 8 {
			token.Prefix = "kc_****" + tokenHash[:8]
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		logger.Error("[DATABASE] Error for api token: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tokens, nil
}

// RevokeAPIToken deactivates a token (soft delete)
func (d *Database) RevokeAPIToken(userID int, tokenID int) error {
	logger.Debug("[DATABASE] Begin RevokeAPIToken(userID:%d, tokenID:%d)", userID, tokenID)

	query := `
		UPDATE api_tokens
		SET is_active = false
		WHERE id = $1 AND user_id = $2
	`

	result, err := d.db.Exec(query, tokenID, userID)
	if err != nil {
		logger.Error("[DATABASE] Error updating api token: %v", err)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already revoked")
	}

	return nil
}

// DeleteAPIToken permanently deletes a token
func (d *Database) DeleteAPIToken(userID int, tokenID int) error {
	logger.Debug("[DATABASE] Begin DeleteAPIToken(userID:%d, tokenID:%d)", userID, tokenID)

	query := `
		DELETE FROM api_tokens
		WHERE id = $1 AND user_id = $2
	`

	result, err := d.db.Exec(query, tokenID, userID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting api token: %v", err)
		return fmt.Errorf("failed to delete token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// GetAPITokenByID retrieves a single token (for display, not the raw token)
func (d *Database) GetAPITokenByID(userID int, tokenID int) (*models.APITokenListResponse, error) {
	logger.Debug("[DATABASE] Begin GetAPITokenByID(userID:%d, tokenID:%d)", userID, tokenID)

	query := `
		SELECT 
			id, 
			name, 
			token_hash,
			last_used_at, 
			created_at, 
			expires_at, 
			is_active,
			CASE 
				WHEN expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP THEN true
				ELSE false
			END as is_expired
		FROM api_tokens
		WHERE id = $1 AND user_id = $2
	`

	var token models.APITokenListResponse
	var lastUsedAt sql.NullTime
	var expiresAt sql.NullTime
	var tokenHash string

	err := d.db.QueryRow(query, tokenID, userID).Scan(
		&token.ID,
		&token.Name,
		&tokenHash,
		&lastUsedAt,
		&token.CreatedAt,
		&expiresAt,
		&token.IsActive,
		&token.IsExpired,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found")
	}
	if err != nil {
		logger.Error("[DATABASE] Error selecting api token: %v", err)
		return nil, fmt.Errorf("failed to fetch token: %w", err)
	}

	// Convert nullable fields
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}

	// Generate prefix from hash for display
	if len(tokenHash) >= 8 {
		token.Prefix = "kc_****" + tokenHash[:8]
	}

	return &token, nil
}
