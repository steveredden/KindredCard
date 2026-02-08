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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

// CreateContact creates a new contact
func (d *Database) CreateContact(userID int, contact *models.Contact) error {
	logger.Debug("[DATABASE] Begin CreateContact(userID:%d, contact:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Contact:")
		utils.Dump(contact)
	}

	tx, err := d.db.Begin()
	if err != nil {
		logger.Error("[DATABASE] Error starting tx: %v", err)
		return err
	}
	defer tx.Rollback()

	// Generate UID if not provided
	if contact.UID == "" {
		contact.UID = fmt.Sprintf("%d-%d", time.Now().Unix(), time.Now().Nanosecond())
	}

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	// Generate ETag
	contact.ETag = fmt.Sprintf("%x", time.Now().UnixNano())

	query := `
		INSERT INTO contacts (uid, full_name, given_name, family_name, middle_name, prefix, suffix, 
			nickname, gender, birthday, birthday_month, birthday_day, anniversary, anniversary_month,
			anniversary_day, notes, avatar_base64, avatar_mime_type, exclude_from_sync, etag, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query,
		contact.UID, contact.FullName, contact.GivenName, contact.FamilyName,
		contact.MiddleName, contact.Prefix, contact.Suffix, contact.Nickname,
		contact.Gender, contact.Birthday, contact.BirthdayMonth, contact.BirthdayDay,
		contact.Anniversary, contact.AnniversaryMonth, contact.AnniversaryDay,
		contact.Notes, contact.AvatarBase64, contact.AvatarMimeType,
		contact.ExcludeFromSync, contact.ETag, userID,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		logger.Error("[DATABASE] Error inserting contact: %v", err)
		return err
	}

	// Insert related data
	if err := d.insertEmails(tx, contact.ID, contact.Emails); err != nil {
		return err
	}
	if err := d.insertPhones(tx, contact.ID, contact.Phones); err != nil {
		return err
	}
	if err := d.insertAddresses(tx, contact.ID, contact.Addresses); err != nil {
		return err
	}
	if err := d.insertOrganizations(tx, contact.ID, contact.Organizations); err != nil {
		return err
	}
	if err := d.insertURLs(tx, contact.ID, contact.URLs); err != nil {
		return err
	}
	if err := d.insertOtherDates(tx, contact.ID, contact.OtherDates); err != nil {
		return err
	}
	if err := d.insertRelationships(tx, contact.ID, contact.Relationships); err != nil {
		return err
	}
	if err := d.insertOtherRelationships(tx, contact.ID, contact.OtherRelationships); err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if err := d.bumpContactSyncToken(contact.ID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return nil
}

// GetAllContactsAbbrv retrieves abbreviated contact information
// scoped to a specific user and optionally filtered by the exclude_from_sync flag.
func (d *Database) GetAllContactsAbbrv(userID int, excludeFromSync bool) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetAllContactsAbbrv(userID:%d, excludeFromSync:%v)", userID, excludeFromSync)

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`SELECT uid, id, full_name, given_name, family_name, nickname, etag FROM contacts WHERE user_id = $1 AND deleted_at IS NULL`)

	params := []interface{}{userID}

	// Add the optional filter for exclude_from_sync
	if excludeFromSync {
		queryBuilder.WriteString(" AND exclude_from_sync != true")
	}

	// Append the sorting and finalize the query string
	queryBuilder.WriteString(" ORDER BY full_name")
	query := queryBuilder.String()

	// Execute the query using the collected arguments
	rows, err := d.db.Query(query, params...)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, fmt.Errorf("error executing query for GetAllContactsAbbrv: %w", err)
	}
	defer rows.Close()

	// Process the results
	contacts := []*models.Contact{}
	for rows.Next() {
		contact := &models.Contact{}
		err := rows.Scan(
			&contact.UID, &contact.ID, &contact.FullName, &contact.GivenName, &contact.FamilyName,
			&contact.Nickname, &contact.ETag,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, fmt.Errorf("error scanning contact row: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		logger.Error("[DATABASE] Error iterating contacts: %v", err)
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return contacts, nil
}

// GetAllContacts retrieves * from all contacts (and related data)
// basically just a wrapper for other calls
func (d *Database) GetAllContacts(userID int, excludeFromSync bool) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetAllContacts(userID:%d, excludeFromSync:%v)", userID, excludeFromSync)

	listContacts, err := d.GetAllContactsAbbrv(userID, excludeFromSync)
	if err != nil {
		return nil, err
	}

	var contacts []*models.Contact
	for _, ix := range listContacts {
		contact, err := d.GetContactByID(userID, ix.ID)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContact retrieves a contact by ID
func (d *Database) GetContactByID(userID int, contactID int) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactByID(userID:%d, contactID:%d)", userID, contactID)

	contact := &models.Contact{}

	var avatarBase64 sql.NullString
	var avatarMimeType sql.NullString
	var family_name sql.NullString
	var gender sql.NullString
	var middle_name sql.NullString
	var nickname sql.NullString
	var maiden_name sql.NullString
	var notes sql.NullString
	var prefix sql.NullString
	var suffix sql.NullString

	var anniversary_day sql.NullInt64
	var anniversary_month sql.NullInt64
	var birthday_day sql.NullInt64
	var birthday_month sql.NullInt64

	var anniversary sql.NullTime
	var birthday sql.NullTime

	query := `
		SELECT id, uid, full_name, given_name, family_name, middle_name, prefix, suffix,
			nickname, maiden_name, gender, birthday, birthday_month, birthday_day, anniversary, 
			anniversary_month, anniversary_day, notes, avatar_base64, avatar_mime_type,
			exclude_from_sync, last_modified_token, created_at, updated_at, etag
		FROM contacts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	err := d.db.QueryRow(query, contactID, userID).Scan(
		&contact.ID, &contact.UID, &contact.FullName, &contact.GivenName, &family_name,
		&middle_name, &prefix, &suffix, &nickname, &maiden_name, &gender, &birthday, &birthday_month,
		&birthday_day, &anniversary, &anniversary_month, &anniversary_day, &notes,
		&avatarBase64, &avatarMimeType, &contact.ExcludeFromSync, &contact.LastModifiedToken,
		&contact.CreatedAt, &contact.UpdatedAt, &contact.ETag,
	)

	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}

	// Load string conversions
	contact.AvatarBase64 = utils.ScanNullString(avatarBase64)
	contact.AvatarMimeType = utils.ScanNullString(avatarMimeType)
	contact.Gender = utils.ScanNullString(gender)
	contact.FamilyName = utils.ScanNullString(family_name)
	contact.MiddleName = utils.ScanNullString(middle_name)
	contact.Nickname = utils.ScanNullString(nickname)
	contact.MaidenName = utils.ScanNullString(maiden_name)
	contact.Notes = utils.ScanNullString(notes)
	contact.Prefix = utils.ScanNullString(prefix)
	contact.Suffix = utils.ScanNullString(suffix)

	// Load int conversions
	contact.AnniversaryDay = utils.ScanNullInt(anniversary_day)
	contact.AnniversaryMonth = utils.ScanNullInt(anniversary_month)
	contact.BirthdayDay = utils.ScanNullInt(birthday_day)
	contact.BirthdayMonth = utils.ScanNullInt(birthday_month)

	// Load time conversions
	contact.Anniversary = utils.ScanNullTime(anniversary)
	contact.Birthday = utils.ScanNullTime(birthday)

	// Load related data
	contact.Emails, _ = d.getEmails(contact.ID)
	contact.Phones, _ = d.getPhones(contact.ID)
	contact.Addresses, _ = d.getAddresses(contact.ID)
	contact.Organizations, _ = d.getOrganizations(contact.ID)
	contact.URLs, _ = d.getURLs(contact.ID)
	contact.OtherDates, _ = d.getOtherDates(contact.ID)
	contact.Relationships, _ = d.getAllRelationships(contact.ID)
	contact.OtherRelationships, _ = d.getOtherRelationships(contact.ID)

	contact.UserID = userID

	return contact, nil
}

// GetContact retrieves a contact by ID
func (d *Database) GetContactByIDAbbrv(userID int, contactID int) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactByIDAbbrv(userID:%d, contactID:%d)", userID, contactID)

	contact := &models.Contact{}

	var family_name sql.NullString
	var gender sql.NullString
	var nickname sql.NullString
	var birthday sql.NullTime
	var birthday_day sql.NullInt64
	var birthday_month sql.NullInt64

	query := `
		SELECT id, uid, full_name, given_name, family_name, nickname, gender, birthday,
			birthday_month, birthday_day, avatar_base64, avatar_mime_type
		FROM contacts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	err := d.db.QueryRow(query, contactID, userID).Scan(
		&contact.ID, &contact.UID, &contact.FullName, &contact.GivenName, &family_name,
		&nickname, &gender, &birthday, &birthday_month, &birthday_day,
		&contact.AvatarBase64, &contact.AvatarMimeType,
	)

	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}

	// Load string conversions
	contact.Gender = utils.ScanNullString(gender)
	contact.FamilyName = utils.ScanNullString(family_name)
	contact.Nickname = utils.ScanNullString(nickname)

	// Load int conversions
	contact.BirthdayDay = utils.ScanNullInt(birthday_day)
	contact.BirthdayMonth = utils.ScanNullInt(birthday_month)

	// Load time conversions
	contact.Birthday = utils.ScanNullTime(birthday)

	contact.UserID = userID

	return contact, nil
}

// GetContactIDByUID retrieves a contact by UID
func (d *Database) GetContactIDByUID(userID int, uid string) (int, bool) {
	logger.Debug("[DATABASE] Begin GetContactIDByUID(userID:%d, uid:%s)", userID, uid)

	contact := &models.Contact{}

	query := `SELECT id FROM contacts WHERE uid = $1 AND deleted_at IS NULL AND user_id = $2`
	err := d.db.QueryRow(query, uid, userID).Scan(&contact.ID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return 1, false
	}

	return contact.ID, true
}

// GetContactByUID retrieves a contact by UID
func (d *Database) GetContactByUID(userID int, uid string, excludeFromSync bool) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactByUID(userID:%d, uid:%s, excludeFromSync:%v)", userID, uid, excludeFromSync)

	contact := &models.Contact{}

	var queryBuilder strings.Builder

	var avatarBase64 sql.NullString
	var avatarMimeType sql.NullString
	var family_name sql.NullString
	var gender sql.NullString
	var middle_name sql.NullString
	var nickname sql.NullString
	var maiden_name sql.NullString
	var notes sql.NullString
	var prefix sql.NullString
	var suffix sql.NullString

	var anniversary_day sql.NullInt64
	var anniversary_month sql.NullInt64
	var birthday_day sql.NullInt64
	var birthday_month sql.NullInt64

	var anniversary sql.NullTime
	var birthday sql.NullTime

	queryBuilder.WriteString(`
		SELECT id, uid, full_name, given_name, family_name, middle_name, prefix, suffix,
			nickname, maiden_name, gender, birthday, birthday_month, birthday_day, anniversary,
			anniversary_month, anniversary_day, notes, avatar_base64, avatar_mime_type,
			exclude_from_sync, last_modified_token, created_at, updated_at, etag
		FROM contacts WHERE uid = $1 AND deleted_at IS NULL AND user_id = $2
	`)

	params := []interface{}{uid, userID}

	if excludeFromSync {
		queryBuilder.WriteString(" AND exclude_from_sync != true")
	}

	queryBuilder.WriteString(" LIMIT 1")

	query := queryBuilder.String()

	err := d.db.QueryRow(query, params...).Scan(
		&contact.ID, &contact.UID, &contact.FullName, &contact.GivenName,
		&family_name, &middle_name, &prefix, &suffix, &nickname, &maiden_name, &gender,
		&birthday, &birthday_month, &birthday_day, &anniversary,
		&anniversary_month, &anniversary_day, &notes, &avatarBase64,
		&avatarMimeType, &contact.ExcludeFromSync, &contact.LastModifiedToken,
		&contact.CreatedAt, &contact.UpdatedAt, &contact.ETag,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}

	// Load string conversions
	contact.AvatarBase64 = utils.ScanNullString(avatarBase64)
	contact.AvatarMimeType = utils.ScanNullString(avatarMimeType)
	contact.Gender = utils.ScanNullString(gender)
	contact.FamilyName = utils.ScanNullString(family_name)
	contact.MiddleName = utils.ScanNullString(middle_name)
	contact.Nickname = utils.ScanNullString(nickname)
	contact.MaidenName = utils.ScanNullString(maiden_name)
	contact.Notes = utils.ScanNullString(notes)
	contact.Prefix = utils.ScanNullString(prefix)
	contact.Suffix = utils.ScanNullString(suffix)

	// Load int conversions
	contact.AnniversaryDay = utils.ScanNullInt(anniversary_day)
	contact.AnniversaryMonth = utils.ScanNullInt(anniversary_month)
	contact.BirthdayDay = utils.ScanNullInt(birthday_day)
	contact.BirthdayMonth = utils.ScanNullInt(birthday_month)

	// Load time conversions
	contact.Anniversary = utils.ScanNullTime(anniversary)
	contact.Birthday = utils.ScanNullTime(birthday)

	// Load related data
	contact.Emails, _ = d.getEmails(contact.ID)
	contact.Phones, _ = d.getPhones(contact.ID)
	contact.Addresses, _ = d.getAddresses(contact.ID)
	contact.Organizations, _ = d.getOrganizations(contact.ID)
	contact.URLs, _ = d.getURLs(contact.ID)
	contact.OtherDates, _ = d.getOtherDates(contact.ID)
	contact.Relationships, _ = d.getAllRelationships(contact.ID)
	contact.OtherRelationships, _ = d.getOtherRelationships(contact.ID)

	return contact, nil
}

// UpdateContact updates an existing contact
func (d *Database) UpdateContact(userID int, contact *models.Contact) error {
	logger.Debug("[DATABASE] Begin UpdateContact(userID:%d, contact:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Contact:")
		utils.Dump(contact)
	}

	tx, err := d.db.Begin()
	if err != nil {
		logger.Error("[DATABASE] Error starting tx: %v", err)
		return err
	}
	defer tx.Rollback()

	// Update ETag
	contact.ETag = fmt.Sprintf("%x", time.Now().UnixNano())

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	query := `
		UPDATE contacts SET
			full_name = $1, given_name = $2, family_name = $3, middle_name = $4, prefix = $5,
			suffix = $6, nickname = $7, maiden_name = $8, gender = $9, birthday = $10, birthday_month = $11,
			birthday_day = $12, anniversary = $13, anniversary_month = $14, anniversary_day = $15, 
			notes = $16, exclude_from_sync = $17, etag = $18
		WHERE id = $19
	`

	_, err = tx.Exec(query,
		contact.FullName, contact.GivenName, contact.FamilyName, contact.MiddleName, contact.Prefix,
		contact.Suffix, contact.Nickname, contact.MaidenName, contact.Gender, contact.Birthday,
		contact.BirthdayMonth, contact.BirthdayDay, contact.Anniversary, contact.AnniversaryMonth,
		contact.AnniversaryDay, contact.Notes, contact.ExcludeFromSync, contact.ETag,
		contact.ID,
	)

	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return err
	}

	// Delete and re-insert related data
	tables := []string{"emails", "phones", "addresses", "organizations", "urls", "other_dates", "other_relationships"}

	// Quick fix for #6 - if saved from the GUI then don't delete and insert relationships
	if contact.Metadata == "skip relationships" {
		logger.Debug("[DATABASE] Skipping relationships rebuild due to GUI PUT")
	} else {
		tables = append(tables, "relationships")
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE contact_id = $1", table)
		if _, err := tx.Exec(query, contact.ID); err != nil {
			logger.Error("[DATABASE] Error deleting from %s: %v", table, err)
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	if err := d.insertEmails(tx, contact.ID, contact.Emails); err != nil {
		return err
	}
	if err := d.insertPhones(tx, contact.ID, contact.Phones); err != nil {
		return err
	}
	if err := d.insertAddresses(tx, contact.ID, contact.Addresses); err != nil {
		return err
	}
	if err := d.insertOrganizations(tx, contact.ID, contact.Organizations); err != nil {
		return err
	}
	if err := d.insertURLs(tx, contact.ID, contact.URLs); err != nil {
		return err
	}
	if err := d.insertOtherDates(tx, contact.ID, contact.OtherDates); err != nil {
		return err
	}
	// Quick fix for #6 - if saved from the GUI then don't delete and insert relationships
	if contact.Metadata != "skip relationships" {
		if err := d.insertRelationships(tx, contact.ID, contact.Relationships); err != nil {
			return err
		}
	}
	if err := d.insertOtherRelationships(tx, contact.ID, contact.OtherRelationships); err != nil {
		return err
	}

	if contact.AvatarBase64 != "" && contact.AvatarMimeType != "" {
		_, err := tx.Exec(`
				UPDATE contacts 
				SET avatar_base64 = $1, avatar_mime_type = $2
				WHERE id = $3`,
			contact.AvatarBase64, contact.AvatarMimeType, contact.ID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if err := d.bumpContactSyncToken(contact.ID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return nil
}

// DeleteContact deletes a contact
func (d *Database) DeleteContact(userID int, contactID int) error {
	logger.Debug("[DATABASE] Begin DeleteContact(userID:%d, contactID:%d)", userID, contactID)

	// 1. Start a transaction to ensure atomicity of the token update and the soft delete.
	tx, err := d.db.Begin()
	if err != nil {
		logger.Error("[DATABASE] Error starting tx: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer a rollback in case of an error later. If Commit succeeds, Rollback is a no-op.
	defer tx.Rollback()

	// --- STEP A: Increment the User's Global Sync Token (Get the new revision number) ---
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	// --- STEP B: Soft-Delete the Contact and Stamp it with the New Token ---

	// Delete any associated records - we can't rely on `ON DELETE CASCADE` if we soft delete...
	tables := []string{"emails", "phones", "addresses", "organizations", "urls", "other_dates", "relationships", "other_relationships"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE contact_id = $1", table)
		if _, err := tx.Exec(query, contactID); err != nil {
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	// We update the contact's row, setting both the deleted_at timestamp
	// and the last_modified_token to the new value.
	contactQuery := `
        UPDATE contacts 
        SET deleted_at = $1, version_token = $2, last_modified_token = $3, etag = $4
        WHERE id = $5 AND user_id = $6`

	// We generate a new ETag to signify a change in the resource state (from present to deleted/gone).
	// Using the token itself or a derivative of the new token is a good practice for the ETag here.
	newETag := fmt.Sprintf("DEL-%d", newSyncToken)

	res, err := tx.Exec(contactQuery, time.Now(), newSyncToken, newSyncToken, newETag, contactID, userID)
	if err != nil {
		logger.Error("[DATABASE] Error soft-deleting contact: %v", err)
		return fmt.Errorf("failed to soft-delete contact: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error("[DATABASE] Error retreiving deleted contact: %v", err)
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		// Return a specific error if the contact was not found or didn't belong to the user
		return sql.ErrNoRows
	}

	// --- STEP C: Commit the Transaction ---
	if err := tx.Commit(); err != nil {
		logger.Error("[DATABASE] Error committing contact tx: %v", err)
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	return nil
}

// SearchContacts searches for contacts by name or email
func (d *Database) SearchContacts(userID int, query string) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin SearchContacts(userID:%d, query:%s)", userID, query)

	var avatarBase64 sql.NullString
	var avatarMimeType sql.NullString
	var family_name sql.NullString
	var middle_name sql.NullString
	var nickname sql.NullString
	var maiden_name sql.NullString
	var notes sql.NullString
	var prefix sql.NullString
	var suffix sql.NullString

	searchQuery := `
		SELECT DISTINCT c.id, c.uid, c.full_name, c.given_name, c.family_name, c.middle_name,
			c.prefix, c.suffix, c.nickname, c.maiden_name, c.birthday, c.anniversary, c.notes, c.avatar_base64,
			c.avatar_mime_type, c.exclude_from_sync, c.created_at, c.updated_at, c.etag
		FROM contacts c
		LEFT JOIN emails e ON c.id = e.contact_id
		WHERE c.user_id = $2 AND c.deleted_at IS NULL AND (
				c.full_name ILIKE $1 
				OR c.given_name ILIKE $1 
				OR c.family_name ILIKE $1 
				OR c.nickname ILIKE $1
				OR c.maiden_name ILIKE $1
				OR e.email ILIKE $1
			)
		ORDER BY c.full_name`

	searchPattern := "%" + query + "%"
	rows, err := d.db.Query(searchQuery, searchPattern, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		contact := &models.Contact{}
		err := rows.Scan(
			&contact.ID, &contact.UID, &contact.FullName, &contact.GivenName, &family_name,
			&middle_name, &prefix, &suffix, &nickname, &maiden_name, &contact.Birthday, &contact.Anniversary,
			&notes, &avatarBase64, &avatarMimeType, &contact.ExcludeFromSync,
			&contact.CreatedAt, &contact.UpdatedAt, &contact.ETag,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, err
		}

		contact.AvatarBase64 = utils.ScanNullString(avatarBase64)
		contact.AvatarMimeType = utils.ScanNullString(avatarMimeType)
		contact.FamilyName = utils.ScanNullString(family_name)
		contact.MiddleName = utils.ScanNullString(middle_name)
		contact.Nickname = utils.ScanNullString(nickname)
		contact.MaidenName = utils.ScanNullString(maiden_name)
		contact.Notes = utils.ScanNullString(notes)
		contact.Prefix = utils.ScanNullString(prefix)
		contact.Suffix = utils.ScanNullString(suffix)

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetContactCount returns the total number of contacts
func (d *Database) GetContactCount(userID int) (int, error) {
	logger.Debug("[DATABASE] Begin GetContactCount(userID:%d)", userID)

	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE user_id = $1 AND deleted_at IS NULL", userID).Scan(&count)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
	}
	return count, err
}

// GetRecentlyEditedCountByDays returns the total number of contacts edited within <dayLookback> days
func (d *Database) GetRecentlyEditedCountByDays(userID int, dayLookback int) (int, error) {
	logger.Debug("[DATABASE] Begin GetRecentlyEditedCountByDays(userID:%d, dayLookback:%d)", userID, dayLookback)

	var count int

	updatedCutoff := time.Now().AddDate(0, 0, -dayLookback)

	err := d.db.QueryRow(`
		SELECT COUNT(*)
		FROM contacts 
		WHERE user_id = $1 AND deleted_at IS NULL AND updated_at >= $2`,
		userID, updatedCutoff).Scan(&count)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
	}
	return count, err
}

// GetRecentlyEditedContacts returns the most recently updated contacts
func (d *Database) GetRecentlyEditedContacts(userID int, limit int) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetRecentlyEditedContacts(userID:%d, limit:%d)", userID, limit)

	var family_name sql.NullString

	rows, err := d.db.Query(`
		SELECT id, full_name, given_name, family_name
		FROM contacts
		WHERE exclude_from_sync = false AND deleted_at IS NULL AND user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		contact := &models.Contact{}
		err := rows.Scan(
			&contact.ID, &contact.FullName, &contact.GivenName, &family_name,
		)
		if err != nil {
			continue
		}

		contact.FamilyName = utils.ScanNullString(family_name)
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// UpdateAvatar updates contact avatars with base64
func (d *Database) UpdateAvatar(userID int, contactID int, avatarBase64 string, mimeType string) error {
	logger.Debug("[DATABASE] Begin UpdateAvatar(userID:%d, contactID:%d, avatarBase64:--, mimeType:%s)", userID, contactID, mimeType)

	_, err := d.db.Exec(`
		UPDATE contacts 
		SET avatar_base64 = $1, avatar_mime_type = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		avatarBase64, mimeType, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error updating avatar: %v", err)
	}

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return err
}

// DeleteAvatar removes contact avatar
func (d *Database) DeleteAvatar(userID int, contactID int) error {
	logger.Debug("[DATABASE] Begin DeleteAvatar(userID:%d, contactID:%d)", userID, contactID)

	_, err := d.db.Exec(`
		UPDATE contacts 
		SET avatar_base64 = NULL, avatar_mime_type = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting avatar: %v", err)
	}

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return err
}

// ListContactsChangedSince fetches all contacts (including soft-deleted ones)
// for a user whose version_token is greater than the client's last known token.
func (d *Database) ListContactsChangedSince(userID int, clientToken int64, excludeFromSync bool) ([]models.Contact, error) {
	logger.Debug("[DATABASE] Begin ListContactsChangedSince(userID:%d, clientToken:%d, excludeFromSync:%v)", userID, clientToken, excludeFromSync)

	contacts := []models.Contact{}

	var queryBuilder strings.Builder

	// CRITICAL CHANGE: We now select the 'deleted_at' column.
	// We do NOT use WHERE deleted_at IS NULL, because we need the deleted records (tombstones).
	queryBuilder.WriteString(`
        SELECT uid, etag, last_modified_token, deleted_at, version_token
        FROM contacts 
        WHERE user_id = $1 AND version_token > $2 
	`)

	params := []interface{}{userID, clientToken}

	if excludeFromSync {
		queryBuilder.WriteString(" AND exclude_from_sync != true")
	}

	queryBuilder.WriteString(" ORDER BY version_token ASC")
	query := queryBuilder.String()

	rows, err := d.db.Query(query, params...)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, fmt.Errorf("error executing query for ListContactsChangedSince: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Contact
		var deletedAt sql.NullTime

		err := rows.Scan(&c.UID, &c.ETag, &c.LastModifiedToken, &deletedAt, &c.VersionToken)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, fmt.Errorf("error scanning contact row: %w", err)
		}

		// Convert sql.NullTime to the pointer *time.Time in the model for easier checking later.
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}

		contacts = append(contacts, c)
	}

	if err := rows.Err(); err != nil {
		logger.Error("[DATABASE] Error iterating contacts: %v", err)
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return contacts, nil
}

func (d *Database) PatchContact(userID int, contactID int, patch *models.ContactJSONPatch) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin PatchContact(userID:%d, contactID:%d)", userID, contactID)

	// Check if there are any updates
	if !patch.HasUpdates() {
		return d.GetContactByID(userID, contactID)
	}

	// Define field mappings
	fieldUpdates := []struct {
		value      interface{}
		columnName string
	}{
		{patch.GivenName, "given_name"},
		{patch.FamilyName, "family_name"},
		{patch.MiddleName, "middle_name"},
		{patch.Prefix, "prefix"},
		{patch.Suffix, "suffix"},
		{patch.Nickname, "nickname"},
		{patch.MaidenName, "maiden_name"},
		{patch.Gender, "gender"},
		{patch.Notes, "notes"},
		{patch.AvatarBase64, "avatar_base64"},
		{patch.AvatarMimeType, "avatar_mime_type"},
		{patch.ExcludeFromSync, "exclude_from_sync"},
	}

	// ensure we can calculate a proper full_name based on this
	current, err := d.GetContactByID(userID, contactID)
	if err != nil {
		return nil, err
	}

	if patch.Prefix != nil {
		current.Prefix = *patch.Prefix
	}
	if patch.GivenName != nil {
		current.GivenName = *patch.GivenName
	}
	if patch.MiddleName != nil {
		current.MiddleName = *patch.MiddleName
	}
	if patch.FamilyName != nil {
		current.FamilyName = *patch.FamilyName
	}
	if patch.Suffix != nil {
		current.Suffix = *patch.Suffix
	}

	newName := current.GenerateFullName()

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	// add the computed full_name
	updates = append(updates, fmt.Sprintf("full_name = $%d", argIndex))
	args = append(args, newName)
	argIndex++

	// Build updates dynamically
	for _, field := range fieldUpdates {
		if field.value != nil {
			// Use reflection to dereference any pointer type
			val := reflect.ValueOf(field.value)
			if val.Kind() == reflect.Ptr && !val.IsNil() {
				updates = append(updates, fmt.Sprintf("%s = $%d", field.columnName, argIndex))
				args = append(args, val.Elem().Interface())
				argIndex++
			}
		}
	}

	// Add WHERE clause parameters
	//    WHERE id = $%d AND user_id = $%d
	args = append(args, contactID, userID)

	// Build and execute query
	query := fmt.Sprintf(`
		UPDATE contacts
		SET %s, updated_at = NOW()
		WHERE id = $%d AND user_id = $%d
	`, strings.Join(updates, ", "), argIndex, argIndex+1)

	result, err := d.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to patch contact: %w", err)
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.GetContactByID(userID, contactID)
}

// Helper functions for related data

func (d *Database) insertEmails(tx *sql.Tx, contactID int, emails []models.Email) error {
	for _, email := range emails {
		_, err := tx.Exec(
			"INSERT INTO emails (contact_id, email, label_type_id, is_primary) VALUES ($1, $2, $3, $4)",
			contactID, email.Email, email.Type, email.IsPrimary,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Emails: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertPhones(tx *sql.Tx, contactID int, phones []models.Phone) error {
	for _, phone := range phones {
		_, err := tx.Exec(
			"INSERT INTO phones (contact_id, phone, label_type_id, is_primary) VALUES ($1, $2, $3, $4)",
			contactID, phone.Phone, phone.Type, phone.IsPrimary,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Phones: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertAddresses(tx *sql.Tx, contactID int, addresses []models.Address) error {
	for _, addr := range addresses {
		_, err := tx.Exec(
			"INSERT INTO addresses (contact_id, street, extended_street, city, state, postal_code, country, label_type_id, is_primary) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
			contactID, addr.Street, addr.ExtendedStreet, addr.City, addr.State, addr.PostalCode, addr.Country, addr.Type, addr.IsPrimary,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Addresses: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertOrganizations(tx *sql.Tx, contactID int, orgs []models.Organization) error {
	for _, org := range orgs {
		_, err := tx.Exec(
			"INSERT INTO organizations (contact_id, name, title, department, is_primary) VALUES ($1, $2, $3, $4, $5)",
			contactID, org.Name, org.Title, org.Department, org.IsPrimary,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Organizations: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertURLs(tx *sql.Tx, contactID int, urls []models.URL) error {
	for _, url := range urls {
		_, err := tx.Exec(
			"INSERT INTO urls (contact_id, url, label_type_id) VALUES ($1, $2, $3)",
			contactID, url.URL, url.Type,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting URLs: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertOtherDates(tx *sql.Tx, contactID int, otherDates []models.OtherDate) error {
	for _, otherDate := range otherDates {
		_, err := tx.Exec(`
			INSERT INTO other_dates (contact_id, event_name, event_date, event_date_month, event_date_day)
			VALUES ($1, $2, $3, $4, $5)`,
			contactID, otherDate.EventName, otherDate.EventDate,
			otherDate.EventDateMonth, otherDate.EventDateDay,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Other Dates: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) insertRelationships(tx *sql.Tx, contactID int, relationships []models.Relationship) error {
	for _, relationship := range relationships {
		relatedID := relationship.RelatedContact.ID
		typeID := relationship.RelationshipType.ID

		// Find what the "Mirror" relationship would be
		// (e.g., if we are adding 'Son', the reverse is 'Father' or 'Mother')
		// We get the current contact's gender to know which reverse name to check for
		var myGender string
		_ = tx.QueryRow("SELECT gender FROM contacts WHERE id = $1", contactID).Scan(&myGender)

		reverseTypeID, err := d.GetReverseRelationshipType(typeID, myGender)

		if err == nil {
			// Check if the mirror already exists
			var exists bool
			tx.QueryRow(`
                SELECT EXISTS(
                    SELECT 1 FROM relationships 
                    WHERE contact_id = $1 AND related_contact_id = $2 AND relationship_type_id = $3
                )`, relatedID, contactID, reverseTypeID).Scan(&exists)

			if exists {
				logger.Debug("[DATABASE] Skipping mirror relationship: %d is already %d of %d", contactID, reverseTypeID, relatedID)
				continue
			}
		}

		// If no mirror exists, insert as normal
		_, err = tx.Exec(`
            INSERT INTO relationships (contact_id, related_contact_id, relationship_type_id)
            VALUES ($1, $2, $3)
            ON CONFLICT (contact_id, related_contact_id, relationship_type_id) DO NOTHING`,
			contactID, relatedID, typeID)

		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) insertOtherRelationships(tx *sql.Tx, contactID int, otherRelationships []models.OtherRelationship) error {
	for _, otherRelationship := range otherRelationships {
		_, err := tx.Exec(`
			INSERT INTO other_relationships (contact_id, related_contact_name, relationship_name)
			VALUES ($1, $2, $3)`,
			contactID, otherRelationship.RelatedContactName, otherRelationship.RelationshipName,
		)
		if err != nil {
			logger.Error("[DATABASE] Error inserting Other Relationships: %v", err)
			return err
		}
	}
	return nil
}

func (d *Database) getEmails(contactID int) ([]models.Email, error) {
	query := `
	SELECT e.id, e.contact_id, e.email, e.label_type_id, l.name as type_label, e.is_primary
	FROM emails e
	JOIN contact_label_types l on e.label_type_id = l.id
	WHERE e.contact_id = $1
	`

	rows, err := d.db.Query(query, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Emails: %v", err)
		return nil, err
	}
	defer rows.Close()

	var emails []models.Email
	for rows.Next() {
		var email models.Email
		if err := rows.Scan(&email.ID, &email.ContactID, &email.Email, &email.Type, &email.TypeLabel, &email.IsPrimary); err != nil {
			logger.Error("[DATABASE] Error scanning Emails: %v", err)
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, nil
}

func (d *Database) getPhones(contactID int) ([]models.Phone, error) {

	query := `
	SELECT p.id, p.contact_id, p.phone, p.label_type_id, l.name as type_label, p.is_primary
    FROM phones p
	JOIN contact_label_types l on p.label_type_id = l.id
	WHERE p.contact_id = $1
	`

	rows, err := d.db.Query(query, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Phones: %v", err)
		return nil, err
	}
	defer rows.Close()

	var phones []models.Phone
	for rows.Next() {
		var phone models.Phone
		if err := rows.Scan(&phone.ID, &phone.ContactID, &phone.Phone, &phone.Type, &phone.TypeLabel, &phone.IsPrimary); err != nil {
			logger.Error("[DATABASE] Error scanning Phones: %v", err)
			return nil, err
		}
		phones = append(phones, phone)
	}
	return phones, nil
}

func (d *Database) getAddresses(contactID int) ([]models.Address, error) {
	query := `
	SELECT a.id, a.contact_id, a.street, a.extended_street, a.city, a.state, a.postal_code, a.country, a.label_type_id, l.name as type_label, a.is_primary
	FROM addresses a
	JOIN contact_label_types l on a.label_type_id = l.id
	WHERE a.contact_id = $1
	`

	rows, err := d.db.Query(query, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Addresses: %v", err)
		return nil, err
	}
	defer rows.Close()

	var addresses []models.Address
	for rows.Next() {
		var addr models.Address
		var extendedStreet sql.NullString
		if err := rows.Scan(&addr.ID, &addr.ContactID, &addr.Street, &extendedStreet, &addr.City, &addr.State, &addr.PostalCode, &addr.Country, &addr.Type, &addr.TypeLabel, &addr.IsPrimary); err != nil {
			logger.Error("[DATABASE] Error scanning Addresses: %v", err)
			return nil, err
		}
		addr.ExtendedStreet = utils.ScanNullString(extendedStreet)
		addresses = append(addresses, addr)
	}
	return addresses, nil
}

func (d *Database) getOrganizations(contactID int) ([]models.Organization, error) {
	rows, err := d.db.Query("SELECT id, contact_id, name, title, department, is_primary FROM organizations WHERE contact_id = $1", contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Organizations: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		if err := rows.Scan(&org.ID, &org.ContactID, &org.Name, &org.Title, &org.Department, &org.IsPrimary); err != nil {
			logger.Error("[DATABASE] Error scanning Organizations: %v", err)
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (d *Database) getURLs(contactID int) ([]models.URL, error) {
	query := `
	SELECT u.id, u.contact_id, u.url, u.label_type_id, l.name as type_label
	FROM urls u
	JOIN contact_label_types l on u.label_type_id = l.id
	WHERE contact_id = $1
	`

	rows, err := d.db.Query(query, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting URLs: %v", err)
		return nil, err
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ID, &url.ContactID, &url.URL, &url.Type, &url.TypeLabel); err != nil {
			logger.Error("[DATABASE] Error scanning URLs: %v", err)
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (d *Database) getOtherDates(contactID int) ([]models.OtherDate, error) {

	var event_date sql.NullTime
	var event_date_day sql.NullInt64
	var event_date_month sql.NullInt64

	rows, err := d.db.Query(`
		SELECT id, contact_id, event_name, event_date, event_date_month, event_date_day
		FROM other_dates
		WHERE contact_id = $1`, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Other Dates: %v", err)
		return nil, err
	}
	defer rows.Close()

	var otherDates []models.OtherDate
	for rows.Next() {
		var otherDate models.OtherDate
		if err := rows.Scan(
			&otherDate.ID, &otherDate.ContactID, &otherDate.EventName, &event_date,
			&event_date_month, &event_date_day,
		); err != nil {
			logger.Error("[DATABASE] Error scanning Other Dates: %v", err)
			return nil, err
		}

		otherDate.EventDate = utils.ScanNullTime(event_date)
		otherDate.EventDateMonth = utils.ScanNullInt(event_date_month)
		otherDate.EventDateDay = utils.ScanNullInt(event_date_day)

		otherDates = append(otherDates, otherDate)
	}
	return otherDates, nil
}

func (d *Database) getOtherRelationships(contactID int) ([]models.OtherRelationship, error) {

	rows, err := d.db.Query(`
		SELECT id, contact_id, related_contact_name, relationship_name, created_at
		FROM other_relationships
		WHERE contact_id = $1`, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Other Dates: %v", err)
		return nil, err
	}
	defer rows.Close()

	var otherRelationships []models.OtherRelationship
	for rows.Next() {
		var otherRelationship models.OtherRelationship
		if err := rows.Scan(
			&otherRelationship.ID, &otherRelationship.ContactID, &otherRelationship.RelatedContactName,
			&otherRelationship.RelationshipName, &otherRelationship.CreatedAt,
		); err != nil {
			logger.Error("[DATABASE] Error scanning Other Dates: %v", err)
			return nil, err
		}

		otherRelationships = append(otherRelationships, otherRelationship)
	}
	return otherRelationships, nil
}

func (d *Database) getAllRelationships(contactID int) ([]models.Relationship, error) {
	rows, err := d.db.Query(`
		SELECT r.id,
		       r.contact_id AS contact_id,
		       r.related_contact_id AS related_contact_id,
			   r.created_at AS created_at,
		       rt.name AS relationship_name,
		       rt.reverse_name_male AS reverse_name_male,
		       rt.reverse_name_female AS reverse_name_female,
		       rt.reverse_name_neutral AS reverse_name_neutral,
		       c.full_name AS related_full_name,
		       c.gender AS related_gender,
		       false AS is_reverse
		FROM relationships r
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		JOIN contacts c ON r.related_contact_id = c.id
		WHERE r.contact_id = $1

		UNION ALL

		SELECT r.id,
		       r.related_contact_id AS contact_id,
		       r.contact_id AS related_contact_id,
			   r.created_at AS created_at,
		       rt.name AS relationship_name,
		       rt.reverse_name_male AS reverse_name_male,
		       rt.reverse_name_female AS reverse_name_female,
		       rt.reverse_name_neutral AS reverse_name_neutral,
		       c.full_name AS related_full_name,
		       c.gender AS related_gender,
		       true AS is_reverse
		FROM relationships r
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		JOIN contacts c ON r.contact_id = c.id
		WHERE r.related_contact_id = $1

		ORDER BY relationship_name
	`, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Relationships: %v", err)
		return nil, err
	}
	defer rows.Close()

	var relationships []models.Relationship
	for rows.Next() {
		var rel models.Relationship
		rel.RelationshipType = &models.RelationshipType{}
		rel.RelatedContact = &models.Contact{}

		var isReverse bool
		var relatedGender sql.NullString

		err := rows.Scan(
			&rel.ID,
			&rel.ContactID,
			&rel.RelatedContactID,
			&rel.CreatedAt,
			&rel.RelationshipType.Name,
			&rel.RelationshipType.ReverseNameMale,
			&rel.RelationshipType.ReverseNameFemale,
			&rel.RelationshipType.ReverseNameNeutral,
			&rel.RelatedContact.FullName,
			&relatedGender,
			&isReverse,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning Relationships: %v", err)
			return nil, err
		}

		// If this is a reverse relationship, use the appropriate reverse name based on related contact's gender
		if isReverse {
			gender := ""
			if relatedGender.Valid {
				gender = relatedGender.String
			}

			switch gender {
			case "M", "male":
				rel.RelationshipType.Name = rel.RelationshipType.ReverseNameMale
			case "F", "female":
				rel.RelationshipType.Name = rel.RelationshipType.ReverseNameFemale
			default:
				rel.RelationshipType.Name = rel.RelationshipType.ReverseNameNeutral
			}
		}

		relationships = append(relationships, rel)
	}

	return relationships, nil
}

func (d *Database) bumpContactSyncToken(contactID int, token int) error {
	newETag := fmt.Sprintf("%x", time.Now().UnixNano())
	currentTimestamp := int(time.Now().Unix())

	query := `
		UPDATE contacts SET
			version_token = $1,
			last_modified_token = $2,
			etag = $3
		WHERE id = $4
	`

	_, err := d.db.Exec(query, token, currentTimestamp, newETag, contactID)

	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return err
	}

	return nil
}

// GetContactsByURL retrieves contacts that have a matching URL
func (d *Database) GetContactsByURL(userID int, baseURL string, urlLabelID int) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactsByURL(userID:%d, baseURL: %s, urlLabelID:%d)", userID, baseURL, urlLabelID)

	var birthday sql.NullTime
	var birthday_day sql.NullInt64
	var birthday_month sql.NullInt64
	var avatarBase64 sql.NullString
	var avatarMimeType sql.NullString

	args := []interface{}{}

	query := `
		SELECT 
			c.id, c.uid, c.full_name, c.given_name, c.family_name, c.nickname, c.avatar_base64,
			c.avatar_mime_type, c.birthday, c.birthday_month, c.birthday_day,
			u.id, u.url, u.label_type_id
		FROM urls u
		JOIN contacts c on u.contact_id = c.id
		WHERE c.user_id = $1
	`
	args = append(args, userID)
	argIndex := 2

	if baseURL != "" {
		query += fmt.Sprintf(" and u.url LIKE $%d", argIndex)
		args = append(args, baseURL+"%")
		argIndex++
	}

	if urlLabelID > 0 {
		query += fmt.Sprintf(" and u.label_type_id = $%d", argIndex)
		args = append(args, urlLabelID)
	}

	// Execute the query using the collected arguments
	rows, err := d.db.Query(query, args...)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, fmt.Errorf("error executing query for GetAllContactsAbbrv: %w", err)
	}
	defer rows.Close()

	// Process the results
	contacts := []*models.Contact{}
	for rows.Next() {
		c := &models.Contact{}
		url := &models.URL{}
		err := rows.Scan(
			&c.ID, &c.UID, &c.FullName, &c.GivenName, &c.FamilyName, &c.Nickname, &avatarBase64,
			&avatarMimeType, &birthday, &birthday_month, &birthday_day,
			&url.ID, &url.URL, &url.Type,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, fmt.Errorf("error scanning contact row: %w", err)
		}

		c.AvatarBase64 = utils.ScanNullString(avatarBase64)
		c.AvatarMimeType = utils.ScanNullString(avatarMimeType)

		c.BirthdayDay = utils.ScanNullInt(birthday_day)
		c.BirthdayMonth = utils.ScanNullInt(birthday_month)
		c.Birthday = utils.ScanNullTime(birthday)

		c.URLs = append(c.URLs, *url)

		contacts = append(contacts, c)
	}

	if err := rows.Err(); err != nil {
		logger.Error("[DATABASE] Error iterating contacts: %v", err)
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return contacts, nil
}

// GetAllContactsAbbrv retrieves abbreviated contact information
// scoped to a specific user and optionally filtered by the exclude_from_sync flag.
func (d *Database) GetUnlinkedImmichContacts(userID int) ([]*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetAllContactsAbbrv(userID:%d)", userID)

	immichTypeID, _ := d.GetLabelID("immich", "url")

	query := `
		SELECT 
			c.uid, c.id, c.full_name, c.given_name, c.family_name, c.nickname, c.avatar_base64, c.avatar_mime_type
		FROM contacts c
		WHERE c.user_id = $1 AND c.deleted_at IS NULL
			AND c.id NOT IN (SELECT contact_id FROM urls WHERE label_type_id = $2)
	`

	// Execute the query using the collected arguments
	rows, err := d.db.Query(query, userID, immichTypeID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, fmt.Errorf("error executing query for GetAllContactsAbbrv: %w", err)
	}
	defer rows.Close()

	// Process the results
	contacts := []*models.Contact{}
	for rows.Next() {
		contact := &models.Contact{}

		var avatarBase64 sql.NullString
		var avatarMimeType sql.NullString
		var family_name sql.NullString
		var nickname sql.NullString

		err := rows.Scan(
			&contact.UID, &contact.ID, &contact.FullName, &contact.GivenName, &family_name,
			&nickname, &avatarBase64, &avatarMimeType,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, fmt.Errorf("error scanning contact row: %w", err)
		}

		// Load string conversions
		contact.AvatarBase64 = utils.ScanNullString(avatarBase64)
		contact.AvatarMimeType = utils.ScanNullString(avatarMimeType)
		contact.FamilyName = utils.ScanNullString(family_name)
		contact.Nickname = utils.ScanNullString(nickname)

		// Load related data
		contact.URLs, _ = d.getURLs(contact.ID)

		contact.UserID = userID

		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		logger.Error("[DATABASE] Error iterating contacts: %v", err)
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return contacts, nil
}

// GetLinkedImmichIDs retrieves people IDs from all contacts' URLs associated with this user
func (d *Database) GetLinkedImmichIDs(userID int) (map[string]bool, error) {
	logger.Debug("[DATABASE] Begin GetAlreadyLinkedImmichIDs(userID:%d)", userID)

	immichTypeID, _ := d.GetLabelID("immich", "url")

	query := `
		SELECT u.url 
		FROM urls u
		JOIN contacts c ON u.contact_id = c.id
		WHERE c.user_id = $1 
		AND label_type_id = $2;
	`

	rows, err := d.db.Query(query, userID, immichTypeID)
	if err != nil {
		logger.Error("Error querying URLs: %v", err)
		return nil, err
	}
	defer rows.Close()

	linkedMap := make(map[string]bool)
	for rows.Next() {
		var urlStr string
		err := rows.Scan(&urlStr)
		if err != nil {
			logger.Error("Error scanning URLs: %v", err)
			continue
		}
		id := utils.ExtractIDFromImmichURL(urlStr)
		linkedMap[id] = true
	}
	return linkedMap, nil
}

// DeleteOldContacts removes expired contacts from the database
func (d *Database) DeleteOldContacts() error {
	logger.Debug("[DATABASE] Begin CleanupExpiredSessions()")

	query := `DELETE FROM contacts WHERE deleted_at < NOW() - INTERVAL '30 days';`

	result, err := d.db.Exec(query)
	if err != nil {
		logger.Error("[DATABASE] Error cleaning up deleted contacts: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info("[DATABASE] Cleaned up %d deleted contacts", rowsAffected)
	}

	return nil
}

// GetContactAnniversaryByID retrieves a contact by UID
func (d *Database) GetContactAnniversaryByID(userID int, contactID int) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactAnniversaryByID(userID:%d, contactID:%d)", userID, contactID)

	contact := &models.Contact{}

	var anniversary_day sql.NullInt64
	var anniversary_month sql.NullInt64
	var anniversary sql.NullTime

	query := `
		SELECT id, full_name, anniversary, anniversary_month, anniversary_day
		FROM contacts WHERE id = $1 AND deleted_at IS NULL AND user_id = $2
	`

	err := d.db.QueryRow(query, contactID, userID).Scan(
		&contact.ID, &contact.FullName, &anniversary, &anniversary_month, &anniversary_day,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}

	// Load int conversions
	contact.AnniversaryDay = utils.ScanNullInt(anniversary_day)
	contact.AnniversaryMonth = utils.ScanNullInt(anniversary_month)

	// Load time conversions
	contact.Anniversary = utils.ScanNullTime(anniversary)

	return contact, nil
}

// GetContactByUID retrieves a contact by UID
func (d *Database) GetContactAddressesByID(userID int, contactID int) (*models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactByUID(userID:%d, contactID:%d)", userID, contactID)

	contact := &models.Contact{}

	var anniversary_day sql.NullInt64
	var anniversary_month sql.NullInt64
	var anniversary sql.NullTime

	query := `
		SELECT id, anniversary, anniversary_month, anniversary_day
		FROM contacts WHERE id = $1 AND deleted_at IS NULL AND user_id = $2
	`

	err := d.db.QueryRow(query, contactID).Scan(
		&contact.ID, &anniversary, &anniversary_month, &anniversary_day,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	} else if err != nil {
		logger.Error("[DATABASE] Error selecting contacts: %v", err)
		return nil, err
	}

	// Load int conversions
	contact.AnniversaryDay = utils.ScanNullInt(anniversary_day)
	contact.AnniversaryMonth = utils.ScanNullInt(anniversary_month)

	// Load time conversions
	contact.Anniversary = utils.ScanNullTime(anniversary)

	return contact, nil
}

// UpdateContactDate specifically updates formal dates (birthday / anniversary) on a contact
func (d *Database) UpdateContactNotes(userID int, body models.NotesJSONPut) error {
	logger.Debug("[DATABASE] Begin UpdateContactNotes(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of NotesJSONPut:")
		utils.Dump(body)
	}

	// Build and execute query
	query := `
		UPDATE contacts
		SET notes = $1
		WHERE id = $2 AND user_id = $3
	`

	_, err := d.db.Exec(query, body.Notes, body.ContactID, userID)
	if err != nil {
		return err
	}

	// Increment sync token so CardDAV clients pull the change
	newToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err == nil {
		_ = d.bumpContactSyncToken(body.ContactID, newToken)
	}

	return nil
}
