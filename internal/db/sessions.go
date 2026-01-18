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
	"context"
	"database/sql"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

// CreateSession stores a new session in the database
func (d *Database) CreateSession(session *models.Session) error {
	logger.Debug("[DATABASE] Begin CreateSession(session:--)")

	query := `
		INSERT INTO sessions (
			user_id, token, user_agent, browser, browser_version, 
			os, device, is_mobile, ip_address, referer, language,
			login_time, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	err := d.db.QueryRow(query,
		session.UserID,
		session.Token,
		session.UserAgent,
		session.Browser,
		session.BrowserVer,
		session.OS,
		session.Device,
		session.IsMobile,
		session.IPAddress,
		session.Referer,
		session.Language,
		session.LoginTime,
		session.ExpiresAt,
	).Scan(&session.ID)

	if err != nil {
		logger.Error("[DATABASE] Error creating session: %v", err)
		return err
	}

	return nil
}

// GetSessionByToken retrieves a session by its token
func (d *Database) GetSessionByToken(token string) (*models.Session, error) {
	logger.Debug("[DATABASE] Begin GetSessionByToken(token:--)")

	query := `
		SELECT id, user_id, token, user_agent, browser, browser_version,
			os, device, is_mobile, ip_address, referer, language,
			login_time, last_activity, expires_at
		FROM sessions
		WHERE token = $1 AND expires_at > NOW()`

	session := &models.Session{}
	err := d.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.UserAgent,
		&session.Browser,
		&session.BrowserVer,
		&session.OS,
		&session.Device,
		&session.IsMobile,
		&session.IPAddress,
		&session.Referer,
		&session.Language,
		&session.LoginTime,
		&session.LastActivity,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found or expired
		}
		logger.Error("[DATABASE] Error fetching session: %v", err)
		return nil, err
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (d *Database) GetUserSessions(userID int) ([]*models.Session, error) {
	logger.Debug("[DATABASE] Begin GetUserSessions(userID:%d)", userID)

	query := `
		SELECT id, user_id, token, user_agent, browser, browser_version,
			os, device, is_mobile, ip_address, referer, language,
			login_time, last_activity, expires_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_activity DESC`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		logger.Error("[DATABASE] Error fetching user sessions: %v", err)
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Token,
			&session.UserAgent,
			&session.Browser,
			&session.BrowserVer,
			&session.OS,
			&session.Device,
			&session.IsMobile,
			&session.IPAddress,
			&session.Referer,
			&session.Language,
			&session.LoginTime,
			&session.LastActivity,
			&session.ExpiresAt,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning session: %v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (d *Database) UpdateSessionActivity(token string) error {
	logger.Debug("[DATABASE] Begin UpdateSessionActivity(token:--)")

	query := `UPDATE sessions SET last_activity = NOW() WHERE token = $1`

	_, err := d.db.Exec(query, token)
	if err != nil {
		logger.Error("[DATABASE] Error updating session activity: %v", err)
		return err
	}

	return nil
}

// RevokeSession deletes a session (logout)
func (d *Database) RevokeSession(userID int, token string) error {
	logger.Debug("[DATABASE] Begin RevokeSession(userID:%d, token:--)", userID)

	query := `DELETE FROM sessions WHERE user_id = $1 AND token = $2`

	_, err := d.db.Exec(query, userID, token)
	if err != nil {
		logger.Error("[DATABASE] Error revoking session: %v", err)
		return err
	}

	return nil
}

// RevokeSessionByID deletes a session (logout) given its ID
func (d *Database) RevokeSessionByID(userID int, sessionID int) error {
	logger.Debug("[DATABASE] Begin RevokeSessionByID(userID:%d, sessionID:%d)", userID, sessionID)

	query := `DELETE FROM sessions WHERE user_id = $1 AND id = $2`

	_, err := d.db.Exec(query, userID, sessionID)
	if err != nil {
		logger.Error("[DATABASE] Error revoking session: %v", err)
		return err
	}

	return nil
}

// InvalidateAllSessions deletes all sessions for a user
func (d *Database) InvalidateAllSessions(userID int) error {
	logger.Debug("[DATABASE] Begin InvalidateAllSessions(userID:%d)", userID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start a transaction for atomicity
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// A. Delete all current browser sessions
	_, err = tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// B. Deactivate all API tokens (soft revocation)
	_, err = tx.ExecContext(ctx, "UPDATE api_tokens SET is_active = FALSE WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// CleanupExpiredSessions removes expired sessions from the database
func (d *Database) CleanupExpiredSessions() error {
	logger.Debug("[DATABASE] Begin CleanupExpiredSessions()")

	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := d.db.Exec(query)
	if err != nil {
		logger.Error("[DATABASE] Error cleaning up expired sessions: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info("[DATABASE] Cleaned up %d expired sessions", rowsAffected)
	}

	return nil
}
