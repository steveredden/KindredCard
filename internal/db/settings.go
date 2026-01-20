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
	"database/sql"
	"fmt"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// ========================================
// CONTACT MANAGEMENT
// ========================================

// DeleteAllContacts deletes all contacts for a user
func (d *Database) DeleteAllContacts(userID int) error {

	listContacts, err := d.GetAllContactsAbbrv(userID, false)
	if err != nil {
		return err
	}

	for _, contact := range listContacts {
		d.DeleteContact(userID, contact.ID)
	}

	return nil
}

// FindDuplicateContacts finds potential duplicate contacts
func (d *Database) FindDuplicateContacts(userID int) ([]models.DuplicateGroup, error) {
	logger.Debug("[DATABASE] Begin FindDuplicateContacts(userID:%d)", userID)

	// Find contacts with same name or same email
	query := `
		WITH potential_dupes AS (
			SELECT 
				c1.id as contact1_id,
				c1.given_name || ' ' || COALESCE(c1.family_name, '') as name1,
				c2.id as contact2_id,
				c2.given_name || ' ' || COALESCE(c2.family_name, '') as name2,
				CASE 
					WHEN LOWER(c1.given_name) = LOWER(c2.given_name) 
					     AND LOWER(COALESCE(c1.family_name, '')) = LOWER(COALESCE(c2.family_name, ''))
					THEN 'name'
					ELSE 'email'
				END as match_type
			FROM contacts c1
			JOIN contacts c2 ON c1.user_id = c2.user_id AND c1.id < c2.id
			LEFT JOIN contact_emails e1 ON e1.contact_id = c1.id
			LEFT JOIN contact_emails e2 ON e2.contact_id = c2.id
			WHERE c1.user_id = $1
			AND (
				(LOWER(c1.given_name) = LOWER(c2.given_name) 
				 AND LOWER(COALESCE(c1.family_name, '')) = LOWER(COALESCE(c2.family_name, '')))
				OR (e1.email IS NOT NULL AND LOWER(e1.email) = LOWER(e2.email))
			)
		)
		SELECT DISTINCT contact1_id, name1, contact2_id, name2, match_type
		FROM potential_dupes
		ORDER BY contact1_id
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}
	defer rows.Close()

	duplicates := []models.DuplicateGroup{}
	for rows.Next() {
		var dup models.DuplicateGroup
		err := rows.Scan(&dup.Contact1ID, &dup.Contact1Name, &dup.Contact2ID, &dup.Contact2Name, &dup.MatchType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate: %w", err)
		}
		duplicates = append(duplicates, dup)
	}

	return duplicates, nil
}

// ========================================
// NOTIFICATION SETTINGS
// ========================================

// GetNotificationSettings returns all notification settings for all users
func (d *Database) GetAllNotificationSettings(enabled bool) ([]models.NotificationSetting, error) {
	logger.Debug("[DATABASE] Begin GetAllNotificationSettings(enabled:%v)", enabled)

	query := `
		SELECT 
			id, user_id, name, provider_type, webhook_url, target_address, days_look_ahead,
			notification_time, include_birthdays, include_anniversaries, include_event_dates,
			other_event_regex, enabled, last_sent_at, created_at, updated_at
		FROM notification_settings
	`
	if enabled {
		query += " WHERE enabled = true"
	}

	rows, err := d.db.Query(query)
	if err != nil {
		logger.Error("[DATABASE] Error selecting notification settings: %v", err)
		return nil, fmt.Errorf("failed to get notification settings: %w", err)
	}
	defer rows.Close()

	settings := []models.NotificationSetting{}
	for rows.Next() {
		var lastSentAt sql.NullTime
		var s models.NotificationSetting
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Name, &s.ProviderType, &s.WebhookURL, &s.TargetAddress, &s.DaysLookAhead,
			&s.NotificationTime, &s.IncludeBirthdays, &s.IncludeAnniversaries, &s.IncludeEventDates,
			&s.EventRegex, &s.Enabled, &lastSentAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning notification settings: %v", err)
			return nil, fmt.Errorf("failed to scan notification setting: %w", err)
		}

		// Convert nullable fields
		if lastSentAt.Valid {
			s.LastSentAt = &lastSentAt.Time
		}

		settings = append(settings, s)
	}

	return settings, nil
}

// GetNotificationSettings returns all notification settings for a user
func (d *Database) GetAllUserNotificationSettings(userID int) ([]models.NotificationSetting, error) {
	logger.Debug("[DATABASE] Begin GetAllUserNotificationSettings(userID:%d)", userID)

	query := `
		SELECT 
			id, user_id, name, provider_type, webhook_url, target_address, days_look_ahead,
			notification_time, include_birthdays, include_anniversaries, include_event_dates,
			other_event_regex, enabled, last_sent_at, created_at, updated_at
		FROM notification_settings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting notification settings: %v", err)
		return nil, fmt.Errorf("failed to get notification settings: %w", err)
	}
	defer rows.Close()

	settings := []models.NotificationSetting{}
	for rows.Next() {
		var s models.NotificationSetting
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Name, &s.ProviderType, &s.WebhookURL, &s.TargetAddress, &s.DaysLookAhead,
			&s.NotificationTime, &s.IncludeBirthdays, &s.IncludeAnniversaries, &s.IncludeEventDates,
			&s.EventRegex, &s.Enabled, &s.LastSentAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning notification settings: %v", err)
			return nil, fmt.Errorf("failed to scan notification setting: %w", err)
		}
		settings = append(settings, s)
	}

	return settings, nil
}

// GetNotificationSetting returns a single notification setting
func (d *Database) GetNotificationSettingByID(userID int, notifierID int) (*models.NotificationSetting, error) {
	logger.Debug("[DATABASE] Begin GetNotificationSettingByID(userID:%d, notifierID:%d)", userID, notifierID)

	query := `
		SELECT 
			id, user_id, name, provider_type, webhook_url, target_address, days_look_ahead,
			notification_time, include_birthdays, include_anniversaries, include_event_dates,
			other_event_regex, enabled, last_sent_at, created_at, updated_at
		FROM notification_settings
		WHERE id = $1 AND user_id = $2
	`
	var s models.NotificationSetting
	err := d.db.QueryRow(query, notifierID, userID).Scan(
		&s.ID, &s.UserID, &s.Name, &s.ProviderType, &s.WebhookURL, &s.TargetAddress, &s.DaysLookAhead,
		&s.NotificationTime, &s.IncludeBirthdays, &s.IncludeAnniversaries, &s.IncludeEventDates,
		&s.EventRegex, &s.Enabled, &s.LastSentAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		logger.Error("[DATABASE] Error scanning notification settings: %v", err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification setting not found")
		}
		return nil, fmt.Errorf("failed to get notification setting: %w", err)
	}

	return &s, nil
}

// CreateNotificationSetting creates a new notification setting
func (d *Database) CreateNotificationSetting(userID int, notifier *models.NotificationSetting) (int, error) {
	logger.Debug("[DATABASE] Begin CreateNotificationSetting(userID:%d, notifier:--)", userID)

	query := `
		INSERT INTO notification_settings (
			user_id, name, provider_type, webhook_url, target_address, days_look_ahead,
			notification_time, include_birthdays, include_anniversaries, include_event_dates,
			other_event_regex, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING id
	`

	var id int
	err := d.db.QueryRow(
		query,
		userID, notifier.Name, notifier.ProviderType, notifier.WebhookURL, notifier.TargetAddress, notifier.DaysLookAhead,
		notifier.NotificationTime, notifier.IncludeBirthdays, notifier.IncludeAnniversaries, notifier.IncludeEventDates,
		notifier.EventRegex, notifier.Enabled,
	).Scan(&id)

	if err != nil {
		logger.Error("[DATABASE] Error inserting notification settings: %v", err)
		return 0, fmt.Errorf("failed to create notification setting: %w", err)
	}

	return id, nil
}

// UpdateNotificationSetting updates a notification setting
func (d *Database) UpdateNotificationSetting(userID int, notifier *models.NotificationSetting) error {
	logger.Debug("[DATABASE] Begin UpdateNotificationSetting(userID:%d, notifier:--)", userID)

	query := `
		UPDATE notification_settings
		SET 
			name = $1,
			provider_type = $2,
			webhook_url = $3,
			target_address = $4
			days_look_ahead = $5,
			notification_time = $6,
			include_birthdays = $7,
			include_anniversaries = $8,
			include_event_dates = $9,
			other_event_regex = $10,
			enabled = $11,
			updated_at = NOW()
		WHERE id = $12 AND user_id = $13
	`

	result, err := d.db.Exec(
		query,
		notifier.Name, notifier.ProviderType, notifier.WebhookURL, notifier.TargetAddress, notifier.DaysLookAhead,
		notifier.NotificationTime, notifier.IncludeBirthdays, notifier.IncludeAnniversaries, notifier.IncludeEventDates,
		notifier.EventRegex, notifier.Enabled, notifier.ID,
		userID,
	)
	if err != nil {
		logger.Error("[DATABASE] Error updating notification settings: %v", err)
		return fmt.Errorf("failed to update notification setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("notification setting not found")
	}

	return nil
}

// RecordNotificationSettingSent updates a notification setting `last_sent_at`
func (d *Database) RecordNotificationSettingSent(notifier models.NotificationSetting) error {
	logger.Debug("[DATABASE] Begin RecordNotificationSettingSent(notifier:--)")

	query := `
		UPDATE notification_settings
		SET 
			last_sent_at = NOW()
		WHERE id = $1 AND user_id = $2
	`
	result, err := d.db.Exec(
		query,
		notifier.ID, notifier.UserID,
	)
	if err != nil {
		logger.Error("[DATABASE] Error updating notification settings: %v", err)
		return fmt.Errorf("failed to update notification setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("notification setting not found")
	}

	return nil
}

func (d *Database) HasNotificationBeenSent(notifier models.NotificationSetting) bool {
	logger.Debug("[DATABASE] Begin HasNotificationBeenSent(notifier:--)")

	var count int64

	today := time.Now().Format("2006-01-02")

	query := `
		SELECT COUNT(*) 
		FROM notification_settings
		WHERE id = $1
		AND DATE(last_sent_at) = $2
	`
	err := d.db.QueryRow(query, notifier.ID, today).Scan(&count)

	if err != nil {
		logger.Error("[DATABASE] Error selecting notification settings: %v", err)
		return false // Assume not sent on error to avoid missing notifications
	}

	return count > 0
}

// DeleteNotificationSetting deletes a notification setting
func (d *Database) DeleteNotificationSetting(userID int, notifierID int) error {
	logger.Debug("[DATABASE] Begin DeleteNotificationSetting(userID:%d, token:--)", userID)

	query := `DELETE FROM notification_settings WHERE id = $1 AND user_id = $2`

	result, err := d.db.Exec(query, notifierID, userID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting notification settings: %v", err)
		return fmt.Errorf("failed to delete notification setting: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("notification setting not found")
	}

	return nil
}

// ========================================
// PASSWORD & ACCOUNT MANAGEMENT
// ========================================

// GetUserPreferences returns all preferences for a user
func (d *Database) GetUserPreferences(userID int) (*models.User, error) {
	logger.Debug("[DATABASE] Begin GetUserPreferences(userID:%d)", userID)

	var userPrefs models.User

	query := `
		SELECT theme
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	err := d.db.QueryRow(query, userID).Scan(&userPrefs.Theme)

	if err != nil {
		logger.Error("[DATABASE] Error selecting user preferences: %v", err)
		return nil, err // Assume not sent on error to avoid missing notifications
	}

	return &userPrefs, nil
}

// UpdatePassword updates a user's password
func (d *Database) UpdatePassword(userID int, newPassword string) error {
	logger.Debug("[DATABASE] Begin UpdatePassword(userID:%d, newPassword:--)", userID)

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("[BCRYPT] Error hashing: %v", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = d.db.Exec(`
		UPDATE users 
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, string(hashedPassword), userID)
	if err != nil {
		logger.Error("[DATABASE] Error updating users: %v", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeleteUser deletes a user and all associated data
func (d *Database) DeleteUser(userID int) error {
	logger.Debug("[DATABASE] Begin DeleteUser(userID:%d)", userID)

	tx, err := d.db.Begin()
	if err != nil {
		logger.Error("[DATABASE] Error starting tx: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Mark contact as soft deleted and also delete all related tables
	// _ = d.DeleteAllContacts(userID)

	// consider letting CASCADE hit them?

	// Delete all user data
	tables := []string{
		"notification_settings",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE ", table)
		if table == "users" {
			query += "id = $1"
		} else {
			query += "user_id = $1"
		}

		_, err = tx.Exec(query, userID)
		if err != nil {
			logger.Error("[DATABASE] Error deleting from %s: %v", table, err)
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	return tx.Commit()
}

func (d *Database) GetContactStats(userID int) (*models.ContactStats, error) {
	logger.Debug("[DATABASE] Begin GetContactStats(userID:%d)", userID)

	stats := &models.ContactStats{}

	// Get total contacts
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM contacts WHERE user_id = $1 AND deleted_at IS NULL
	`, userID).Scan(&stats.TotalContacts)
	if err != nil {
		return nil, fmt.Errorf("failed to get total contacts: %w", err)
	}

	// Get contacts added this month
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM contacts 
		WHERE user_id = $1 
		AND deleted_at IS NULL
		AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`, userID).Scan(&stats.AddedThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts added this month: %w", err)
	}

	// Get contacts with birthdays
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM contacts 
		WHERE user_id = $1 
		AND deleted_at IS NULL
		AND birthday IS NOT NULL
	`, userID).Scan(&stats.WithBirthdays)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts with birthdays: %w", err)
	}

	return stats, nil
}

// UpdateUserPreferences updates user theme and contacts per page
func (d *Database) UpdateUserPreferences(user models.User) error {
	logger.Debug("[DATABASE] Begin UpdateUserPreferences(user:--)")

	_, err := d.db.Exec(`
		UPDATE users 
		SET 
			theme = $1
		WHERE id = $2`,
		user.Theme, user.ID)
	return err
}
