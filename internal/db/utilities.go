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
	"fmt"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

func (d *Database) GetContactsMissingGender(userID int) ([]models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactsMissingGender(userID:%d)", userID)

	query := `SELECT id, full_name, given_name, family_name, avatar_base64, avatar_mime_type
	          FROM contacts
	          WHERE user_id = $1 AND (gender IS NULL OR gender = '') AND deleted_at IS NULL
	          LIMIT 50` // Limit to 50 for page performance

	rows, err := d.db.Query(query, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts missing gender: %v", err)
		return nil, fmt.Errorf("failed to get contacts missing gender: %w", err)
	}
	defer rows.Close()

	contacts := []models.Contact{}
	for rows.Next() {
		var c models.Contact
		err := rows.Scan(
			&c.ID, &c.FullName, &c.GivenName, &c.FamilyName, &c.AvatarBase64, &c.AvatarMimeType,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contacts: %v", err)
			return nil, fmt.Errorf("failed to scan contacts: %w", err)
		}

		contacts = append(contacts, c)
	}

	return contacts, nil

}
