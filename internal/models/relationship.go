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

// RelationshipType represents a type of relationship between contacts
type RelationshipType struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	ReverseNameMale    string `json:"reverse_name_male,omitempty"`
	ReverseNameFemale  string `json:"reverse_name_female,omitempty"`
	ReverseNameNeutral string `json:"reverse_name_neutral,omitempty"`
	IsSystem           bool   `json:"is_system"`
}

// Relationship represents a connection between two contacts
type Relationship struct {
	ID               int               `json:"id"`
	ContactID        int               `json:"contact_id"`
	Contact          *Contact          `json:"contact,omitempty"`
	RelatedContactID int               `json:"related_contact_id"`
	RelatedContact   *Contact          `json:"related_contact,omitempty"`
	RelationshipType *RelationshipType `json:"relationship_type"`
	CreatedAt        time.Time         `json:"created_at"`
}

// Relationship represents a connection between two contacts
type OtherRelationship struct {
	ID                 int       `json:"id"`
	ContactID          int       `json:"contact_id"`
	Contact            *Contact  `json:"contact,omitempty"`
	RelatedContactName string    `json:"related_contact_name,omitempty"`
	RelationshipName   string    `json:"relationship_name"`
	CreatedAt          time.Time `json:"created_at"`
}

// Suggestion defines the proposed action for the UI
type RelationshipSuggestion struct {
	Type               string // "Relationship"
	TargetID           int    // The ID of the person we are ADDING the link to
	TargetName         string
	ProposedID         int // The ID of the person they are related to
	SourceName         string
	RelationshipTypeID int    // The ID for "Brother", "Father", etc.
	ProposedVal        string // The Label (e.g., "Brother")
	Reason             string // Your logic description
}
