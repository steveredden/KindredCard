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
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/steveredden/KindredCard/internal/auth"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New(host, port, user, password, dbname string) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// User and Session Management

// CreateUser creates a new user
func (d *Database) CreateUser(email, passwordHash string) (*models.User, error) {
	logger.Debug("[DATABASE] Begin CreateUser(email:%s, passwordHash:--)", email)

	user := &models.User{
		Email:           email,
		PasswordHash:    passwordHash,
		IsSetupComplete: false,
		Theme:           "system",
	}

	err := d.db.QueryRow(`
		INSERT INTO users (email, password_hash, is_setup_complete, theme)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		email, passwordHash, user.IsSetupComplete, user.Theme,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logger.Error("[DATABASE] Error creating user: %v", err)
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(email string) (*models.User, error) {
	logger.Debug("[DATABASE] Begin GetUserByEmail(email:%s)", email)

	user := &models.User{}
	err := d.db.QueryRow(`
		SELECT id, email, password_hash, is_setup_complete, theme, created_at, updated_at
		FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsSetupComplete,
		&user.Theme, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		logger.Error("[DATABASE] Error getting user by email: %v", err)
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func (d *Database) GetUserByID(userID int) (*models.User, error) {
	logger.Debug("[DATABASE] Begin GetUserByID(userID:%d)", userID)

	user := &models.User{}
	err := d.db.QueryRow(`
		SELECT id, email, password_hash, is_setup_complete, theme, created_at, updated_at
		FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsSetupComplete,
		&user.Theme, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		logger.Error("[DATABASE] Error selecting user by ID: %v", err)
		return nil, err
	}
	return user, nil
}

// ValidateUserCredentials checks email and password
func (d *Database) ValidateUserCredentials(email, password string) (*models.User, error) {
	logger.Debug("[DATABASE] Begin ValidateUserCredentials(userID:%s, password:--)", email)

	user, err := d.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		logger.Error("[DATABASE] Error validating user credentials: %v", err)
		return nil, err
	}

	return user, nil
}

// UpdatePasswordHash updates a user's password hash in the database.
func (d *Database) UpdatePasswordHash(userID int, newHash string) error {
	logger.Debug("[DATABASE] Begin UpdatePasswordHash(userID:%d, newHash:--)", userID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        UPDATE users
        SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2`

	result, err := d.db.ExecContext(ctx, query, newHash, userID)
	if err != nil {
		logger.Error("[DATABASE] Error updating user password hash: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows // Indicate that no user was found/updated
	}

	return nil
}

// MarkSetupComplete marks setup as complete for a user
func (d *Database) MarkSetupComplete(userID int) error {
	logger.Debug("[DATABASE] Begin MarkSetupComplete(userID:%d)", userID)

	_, err := d.db.Exec(`
		UPDATE users SET is_setup_complete = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		userID)
	if err != nil {
		logger.Error("[DATABASE] Error marking user setup as complete: %v", err)
	}
	return err
}

// IsSetupComplete checks if any user has completed setup
func (d *Database) IsSetupComplete() (bool, error) {
	logger.Debug("[DATABASE] Begin IsSetupComplete")

	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM users WHERE is_setup_complete = true
	`).Scan(&count)

	return count > 0, err
}

// GetAddressBookSyncToken fetches the latest sync token (BIGINT) for a given user ID.
// This is used to construct the D:sync-token URL in the PROPFIND and REPORT responses.
func (d *Database) GetAddressBookSyncToken(userID int) (int, error) {
	logger.Debug("[DATABASE] Begin GetAddressBookSyncToken(userID:%d)", userID)

	var token int
	// NOTE: Replace 'addressbook_sync_token' with your actual column name
	query := `SELECT addressbook_sync_token FROM users WHERE id = $1`

	err := d.db.QueryRow(query, userID).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // User found, but token not yet set (return default 0)
		}
		logger.Error("[DATABASE] Error selecting user: %v", err)
		return 0, err
	}
	return token, nil
}

// IncrementAndGetNewSyncToken atomically increments the user's sync token and returns the new value.
// This is critical for preventing race conditions during contact creation/update.
func (d *Database) IncrementAndGetNewSyncToken(userID int) (int, error) {
	logger.Debug("[DATABASE] Begin IncrementAndGetNewSyncToken(userID:%d)", userID)

	// 1. Begin a transaction
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // Rollback if not committed

	// 2. Increment the token and return the new value (RETURNING is PostgreSQL syntax)
	// For SQLite or MySQL, you might need two separate UPDATE/SELECT statements within the transaction.
	var newToken int
	query := `
        UPDATE users 
        SET addressbook_sync_token = addressbook_sync_token + 1 
        WHERE id = $1
        RETURNING addressbook_sync_token`

	err = tx.QueryRow(query, userID).Scan(&newToken)
	if err != nil {
		logger.Error("[DATABASE] Error selecting users: %v", err)
		return 0, err
	}

	// 3. Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.Error("[DATABASE] Error commiting user tx: %v", err)
		return 0, err
	}

	return newToken, nil
}
