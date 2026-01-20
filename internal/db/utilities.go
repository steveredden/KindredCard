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

func (d *Database) GetContactsWithPhones(userID int) ([]models.Contact, error) {
	logger.Debug("[DATABASE] Begin GetContactsWithPhones(userID:%d)", userID)

	query := `SELECT c.id, c.full_name, c.given_name, c.family_name, p.id, p.phone 
          FROM phones p
          JOIN contacts c on c.id = p.contact_id
          WHERE c.user_id = $1
          ORDER BY p.last_formatted_at ASC NULLS FIRST, c.full_name ASC 
          LIMIT 50`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting contacts with phones: %v", err)
		return nil, fmt.Errorf("failed to get contacts with phones: %w", err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var c models.Contact
		var p models.Phone // Temporary struct for the phone row

		// Scan both Contact and the specific Phone row
		err := rows.Scan(
			&c.ID, &c.FullName, &c.GivenName, &c.FamilyName,
			&p.ID, &p.Phone,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning contact/phone: %v", err)
			continue
		}

		// Attach the single phone from this row to the contact
		c.Phones = []models.Phone{p}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// Internal struct for processing logic
type internalRel struct {
	ContactID        int
	ContactName      string
	ContactGender    string
	RelatedContactID int
	RelatedName      string
	RelatedGender    string
	RelName          string
}

func (d *Database) GetRelationshipSuggestions(userID int) ([]models.RelationshipSuggestion, error) {
	var suggestions []models.RelationshipSuggestion

	// 1. Preparation
	relTypes, _ := d.GetRelationshipTypes()
	typeMap := make(map[string]int)
	for _, rt := range relTypes {
		typeMap[rt.Name] = rt.ID
	}

	rows, err := d.db.Query(`
		SELECT r.contact_id, c1.full_name, c1.gender, r.related_contact_id, c2.full_name, c2.gender, rt.name
		FROM relationships r
		JOIN contacts c1 ON r.contact_id = c1.id
		JOIN contacts c2 ON r.related_contact_id = c2.id
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		WHERE c1.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type person struct {
		ID     int
		Name   string
		Gender string
	}
	parentsOf := make(map[int][]person)
	childrenOf := make(map[int][]person)
	spousesOf := make(map[int][]person)
	allPeople := make(map[int]person)

	for rows.Next() {
		var r internalRel
		rows.Scan(&r.ContactID, &r.ContactName, &r.ContactGender, &r.RelatedContactID, &r.RelatedName, &r.RelatedGender, &r.RelName)

		p1 := person{r.ContactID, r.ContactName, r.ContactGender}
		p2 := person{r.RelatedContactID, r.RelatedName, r.RelatedGender}
		allPeople[p1.ID] = p1
		allPeople[p2.ID] = p2

		// Indexing (Bidirectional)
		switch r.RelName {
		case "Son", "Daughter", "Child":
			childrenOf[p1.ID] = append(childrenOf[p1.ID], p2)
			parentsOf[p2.ID] = append(parentsOf[p2.ID], p1)
		case "Father", "Mother", "Parent":
			parentsOf[p1.ID] = append(parentsOf[p1.ID], p2)
			childrenOf[p2.ID] = append(childrenOf[p2.ID], p1)
		case "Spouse", "Wife", "Husband":
			spousesOf[p1.ID] = append(spousesOf[p1.ID], p2)
			spousesOf[p2.ID] = append(spousesOf[p2.ID], p1)
		}
	}

	// 2. Inference Engine
	suggestedPairs := make(map[string]bool) // Key: "minID-maxID-Category"

	for id, p := range allPeople {

		// SCENARIO 1: SPOUSE -> CHILD (Step-parents)
		for _, spouse := range spousesOf[id] {
			for _, kid := range childrenOf[spouse.ID] {
				pairKey := fmt.Sprintf("%d-%d-step", id, kid.ID)
				exists, _ := d.RelationExists(id, kid.ID)
				if !exists && id != kid.ID && !suggestedPairs[pairKey] {
					role := inferRole(kid.Gender, "child")
					suggestions = append(suggestions, models.RelationshipSuggestion{
						Type: "Relationship", TargetID: kid.ID, TargetName: kid.Name,
						ProposedID: id, SourceName: p.Name, ProposedVal: role, RelationshipTypeID: typeMap[role],
						Reason: fmt.Sprintf("%s is the spouse of %s, who is the parent of %s.", p.Name, spouse.Name, kid.Name),
					})
					suggestedPairs[pairKey] = true
				}
			}
		}

		// SCENARIO 2: SIBLING INFERENCE
		for _, parent := range parentsOf[id] {
			for _, sib := range childrenOf[parent.ID] {
				if sib.ID == id {
					continue
				}
				// Sort IDs to ensure Mateo/Lucas only appears once
				idA, idB := id, sib.ID
				if idA > idB {
					idA, idB = idB, idA
				}
				pairKey := fmt.Sprintf("%d-%d-sibling", idA, idB)

				exists, _ := d.RelationExists(id, sib.ID)
				if !exists && !suggestedPairs[pairKey] {
					role := inferRole(sib.Gender, "sibling")
					suggestions = append(suggestions, models.RelationshipSuggestion{
						Type: "Relationship", TargetID: sib.ID, TargetName: sib.Name,
						ProposedID: id, SourceName: p.Name, ProposedVal: role, RelationshipTypeID: typeMap[role],
						Reason: fmt.Sprintf("%s and %s are both children of %s.", p.Name, sib.Name, parent.Name),
					})
					suggestedPairs[pairKey] = true
				}
			}
		}

		// SCENARIO 3: GRANDPARENT INFERENCE (Path: A -> B -> C)
		for _, child := range childrenOf[id] {
			for _, gc := range childrenOf[child.ID] {
				pairKey := fmt.Sprintf("%d-%d-grand", id, gc.ID)
				exists, _ := d.RelationExists(id, gc.ID)
				if !exists && id != gc.ID && !suggestedPairs[pairKey] {
					role := inferRole(gc.Gender, "grandchild")
					suggestions = append(suggestions, models.RelationshipSuggestion{
						Type: "Relationship", TargetID: gc.ID, TargetName: gc.Name,
						ProposedID: id, SourceName: p.Name, ProposedVal: role, RelationshipTypeID: typeMap[role],
						Reason: fmt.Sprintf("%s is the parent of %s, who is the parent of %s.", p.Name, child.Name, gc.Name),
					})
					suggestedPairs[pairKey] = true
				}
			}
		}
	}

	return suggestions, nil
}

func (d *Database) RelationExists(contactID, relatedID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM relationships WHERE (contact_id = $1 AND related_contact_id = $2) OR (contact_id = $2 AND related_contact_id = $1))`
	err := d.db.QueryRow(query, contactID, relatedID).Scan(&exists)
	return exists, err
}

func inferRole(gender string, category string) string {
	switch category {
	case "child":
		if gender == "M" {
			return "Son"
		}
		if gender == "F" {
			return "Daughter"
		}
		return "Child"
	case "sibling":
		if gender == "M" {
			return "Brother"
		}
		if gender == "F" {
			return "Sister"
		}
		return "Sibling"
	case "grandchild":
		if gender == "M" {
			return "Grandson"
		}
		if gender == "F" {
			return "Granddaughter"
		}
		return "Grandchild"
	case "grandparent":
		if gender == "M" {
			return "Grandfather"
		}
		if gender == "F" {
			return "Grandmother"
		}
		return "Grandparent"
	default:
		return "Related"
	}
}

func (d *Database) GetAnniversarySuggestions(userID int) ([]models.RelationshipSuggestion, error) {
	var suggestions []models.RelationshipSuggestion

	// 1. Fetch all spouse relationships for this user
	rows, err := d.db.Query(`
		SELECT r.contact_id, c1.full_name, r.related_contact_id, c2.full_name
		FROM relationships r
		JOIN contacts c1 ON r.contact_id = c1.id
		JOIN contacts c2 ON r.related_contact_id = c2.id
		JOIN relationship_types rt ON r.relationship_type_id = rt.id
		WHERE c1.user_id = $1 AND rt.name IN ('Spouse', 'Husband', 'Wife')`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var contactID, relatedID int
		var contactName, relatedName string
		if err := rows.Scan(&contactID, &contactName, &relatedID, &relatedName); err != nil {
			continue
		}

		// 2. Get full contact details to check anniversary fields
		source, _ := d.GetContactByID(userID, contactID)
		target, _ := d.GetContactByID(userID, relatedID)

		// Check Case A: Source has it, Target does not
		if source.HasAnniversary() && !target.HasAnniversary() {
			suggestions = append(suggestions, d.buildAnniversarySuggestion(source, target))
		}

		// Check Case B: Target has it, Source does not
		if target.HasAnniversary() && !source.HasAnniversary() {
			suggestions = append(suggestions, d.buildAnniversarySuggestion(target, source))
		}
	}

	return suggestions, nil
}

// Helper to format the display value and build the suggestion object
func (d *Database) buildAnniversarySuggestion(from *models.Contact, to *models.Contact) models.RelationshipSuggestion {
	var displayValue string
	if from.Anniversary != nil {
		displayValue = from.Anniversary.Format("2006-01-02")
	} else {
		// Send a simple "MM-DD" format even for partials
		displayValue = fmt.Sprintf("%02d-%02d", *from.AnniversaryMonth, *from.AnniversaryDay)
	}

	return models.RelationshipSuggestion{
		Type:        "Event",
		TargetID:    to.ID,
		TargetName:  to.FullName,
		ProposedID:  from.ID,
		SourceName:  from.FullName,
		ProposedVal: displayValue, // This will now be "YYYY-MM-DD" or "MM-DD"
		Reason:      fmt.Sprintf("%s has an anniversary set (%s). Since %s is their spouse, they likely share this date.", from.FullName, displayValue, to.FullName),
	}
}
