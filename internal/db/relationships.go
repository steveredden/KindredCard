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
	"github.com/steveredden/KindredCard/internal/utils"
)

// GetRelationshipTypes retrieves all relationship types
func (d *Database) GetRelationshipTypes() ([]models.RelationshipType, error) {
	logger.Debug("[DATABASE] Begin GetAPITokenByID")

	rows, err := d.db.Query(`SELECT 
		id, name, reverse_name_male, reverse_name_female, reverse_name_neutral, is_system
		FROM relationship_types ORDER BY name`)
	if err != nil {
		logger.Error("[DATABASE] Error selecting relationship types: %v", err)
		return nil, err
	}
	defer rows.Close()

	var types []models.RelationshipType
	for rows.Next() {
		var rt models.RelationshipType
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.ReverseNameMale, &rt.ReverseNameFemale, &rt.ReverseNameNeutral, &rt.IsSystem); err != nil {
			logger.Error("[DATABASE] Error scanning relationship types: %v", err)
			return nil, err
		}
		types = append(types, rt)
	}
	return types, nil
}

func (d *Database) GetReverseRelationshipType(typeID int, gender string) (int, error) {
	logger.Debug("[DATABASE] Begin GetAPITokenByID(typeID:%d, gender:%s)", typeID, gender)

	// Look up the relationship type to find its reverse names
	var name, revMale, revFemale, revNeutral string
	err := d.db.QueryRow("SELECT name, reverse_name_male, reverse_name_female, reverse_name_neutral FROM relationship_types WHERE id = $1", typeID).Scan(&name, &revMale, &revFemale, &revNeutral)
	if err != nil {
		return 0, err
	}

	// Determine which reverse name to look for based on the target's gender
	targetName := revNeutral
	if gender == "M" && revMale != "" {
		targetName = revMale
	} else if gender == "F" && revFemale != "" {
		targetName = revFemale
	}

	if targetName == "" {
		return 0, fmt.Errorf("no reverse type")
	}

	var reverseID int
	err = d.db.QueryRow("SELECT id FROM relationship_types WHERE name = $1", targetName).Scan(&reverseID)
	return reverseID, err
}

func (d *Database) GetRelationshipTypeByName(query string) (models.RelationshipType, error) {
	logger.Debug("[DATABASE] Begin GetRelationshipTypeByName(query:%s)", query)

	var rt models.RelationshipType

	err := d.db.QueryRow(`
		SELECT 
			id, name, reverse_name_male, reverse_name_female, reverse_name_neutral, is_system
		FROM relationship_types
		WHERE name like %$1%
		LIMIT 1`, query).Scan(&rt.ID, &rt.Name, &rt.ReverseNameMale, &rt.ReverseNameFemale, &rt.ReverseNameNeutral, &rt.IsSystem)
	if err != nil {
		logger.Error("[DATABASE] Error selecting relationship types: %v", err)
		return rt, err
	}

	return rt, nil
}

// CreateRelationshipType creates a new custom relationship type
func (d *Database) CreateRelationshipType(relationshipType *models.RelationshipType) (int, error) {
	logger.Debug("[DATABASE] Begin CreateRelationshipType(relationshipType:--)")

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of RelationshipType:")
		utils.Dump(relationshipType)
	}

	var newID int

	err := d.db.QueryRow(`
		INSERT INTO relationship_types (name, reverse_name_male, reverse_name_female, reverse_name_neutral, is_system)
		VALUES ($1, $2, $3)
		RETURNING id`,
		relationshipType.Name, relationshipType.ReverseNameMale, relationshipType.ReverseNameFemale, relationshipType.ReverseNameNeutral,
		false).Scan(&newID)
	if err != nil {
		logger.Error("[DATABASE] Error inserting relationship types: %v", err)
	}

	return newID, err
}

// AddRelationship creates a relationship between two contacts
func (d *Database) AddRelationship(userID int, contactID int, relatedContactID int, relationshipTypeID int) error {
	logger.Debug("[DATABASE] Begin AddRelationship(userID:%d, contactID:%d, relatedContactID:%d, relationshipTypeID:%d)", userID, contactID, relatedContactID, relationshipTypeID)

	// Mirror Detection: Check if the inverse already exists
	var myGender string
	_ = d.db.QueryRow("SELECT gender FROM contacts WHERE id = $1", contactID).Scan(&myGender)

	reverseTypeID, err := d.GetReverseRelationshipType(relationshipTypeID, myGender)
	if err == nil {
		var exists bool
		d.db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM relationships 
                WHERE contact_id = $1 AND related_contact_id = $2 AND relationship_type_id = $3
            )`, relatedContactID, contactID, reverseTypeID).Scan(&exists)

		if exists {
			logger.Info("[DATABASE] Relationship mirror already exists. Skipping redundant insert.")
			return nil
		}
	}

	// Perform the Insert
	_, err = d.db.Exec(`
        INSERT INTO relationships (contact_id, related_contact_id, relationship_type_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (contact_id, related_contact_id, relationship_type_id) DO NOTHING`,
		contactID, relatedContactID, relationshipTypeID)

	if err != nil {
		logger.Error("[DATABASE] Error inserting Relationships: %v", err)
		return err
	}

	// Sync Token Management
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	// Bump both contacts so CardDAV clients see the update for both people
	_ = d.bumpContactSyncToken(contactID, newSyncToken)
	_ = d.bumpContactSyncToken(relatedContactID, newSyncToken)

	return nil
}

// RemoveRelationship removes a relationship between two contacts
func (d *Database) RemoveRelationship(userID int, relationshipID int) error {
	logger.Debug("[DATABASE] Begin RemoveRelationship(userID:%d, relationshipID:%d)", userID, relationshipID)

	rel := &models.Relationship{}

	err := d.db.QueryRow(`
		SELECT contact_id, related_contact_id
		FROM relationships
		WHERE id = $1
		LIMIT 1`, relationshipID).Scan(&rel.ContactID, &rel.RelatedContactID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting Relationship: %v", err)
		return err
	}

	_, err = d.db.Exec("DELETE FROM relationships WHERE id = $1", relationshipID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Relationships: %v", err)
		return err
	}

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(rel.ContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}
	if err := d.bumpContactSyncToken(rel.RelatedContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return nil
}

// RemoveRelationship removes an other_relationship
func (d *Database) RemoveOtherRelationship(userID int, otherRelationshipID int) error {
	logger.Debug("[DATABASE] Begin RemoveOtherRelationship(userID:%d, otherRelationshipID:%d)", userID, otherRelationshipID)

	rel := &models.Relationship{}

	_, err := d.db.Exec("DELETE FROM other_relationships WHERE id = $1", otherRelationshipID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Other Relationships: %v", err)
		return err
	}

	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		logger.Error("[DATABASE] Error incrementing CardDAV sync token: %v", err)
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(rel.ContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}
	if err := d.bumpContactSyncToken(rel.RelatedContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return nil
}
