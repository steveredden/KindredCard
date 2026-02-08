/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package models

import (
	"strings"
	"time"
)

// Contact represents a person in the CRM
type Contact struct {
	ID                     int                 `json:"id" example:"1"`
	UID                    string              `json:"uid" example:"42"`
	FullName               string              `json:"full_name" example:"Dr. John Frank Doe III"` // Computed from name parts
	GivenName              string              `json:"given_name" example:"John"`
	FamilyName             string              `json:"family_name" example:"Doe"`
	MiddleName             string              `json:"middle_name" example:"Frank"`
	Prefix                 string              `json:"prefix" example:"Dr."`
	Suffix                 string              `json:"suffix" example:"III"`
	Nickname               string              `json:"nickname" example:"broheim"`
	MaidenName             string              `json:"maiden_name" example:"Parks"`
	PhoneticFirstName      string              `json:"phonetic_first_name" example:"Par-cor"`
	PronunciationFirstName string              `json:"pronunciation_first_name" example:"Par-cor"`
	PhoneticLastName       string              `json:"phonetic_last_name" example:"Par-cor"`
	PronunciationLastName  string              `json:"pronunciation_last_name" example:"Par-cor"`
	PhoneticMiddleName     string              `json:"phonetic_middle_name" example:"Par-cor"`
	Gender                 string              `json:"gender,omitempty" example:"M"` // M, F, O, N, U
	Birthday               *time.Time          `json:"birthday,omitempty" example:"1990-12-15T00:00:00Z"`
	BirthdayMonth          *int                `json:"birthday_month,omitempty" example:"12"` // 1-12, for partial dates
	BirthdayDay            *int                `json:"birthday_day,omitempty" example:"15"`   // 1-31, for partial dates
	Anniversary            *time.Time          `json:"anniversary,omitempty" example:"2022-01-03T00:00:00Z"`
	AnniversaryMonth       *int                `json:"anniversary_month,omitempty" example:"1"` // 1-12, for partial dates
	AnniversaryDay         *int                `json:"anniversary_day,omitempty" example:"3"`   // 1-31, for partial dates
	Notes                  string              `json:"notes" example:"Met at work conference 2023"`
	AvatarBase64           string              `json:"avatar_base64,omitempty"`
	AvatarMimeType         string              `json:"avatar_mime_type,omitempty"`
	ExcludeFromSync        bool                `json:"exclude_from_sync"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
	ETag                   string              `json:"etag"`
	LastModifiedToken      int64               `json:"last_modified_token"`
	VersionToken           int                 `json:"version_token"`
	UserID                 int                 `json:"user_id"`
	Emails                 []Email             `json:"emails,omitempty"`
	Phones                 []Phone             `json:"phones,omitempty"`
	Addresses              []Address           `json:"addresses,omitempty"`
	Organizations          []Organization      `json:"organizations,omitempty"`
	OtherDates             []OtherDate         `json:"other_dates"`
	URLs                   []URL               `json:"urls,omitempty"`
	Relationships          []Relationship      `json:"relationships,omitempty"`
	OtherRelationships     []OtherRelationship `json:"other_relationships,omitempty"`
	DeletedAt              *time.Time
	Metadata               string
}

// GenerateFullName computes the full name from name components
func (c *Contact) GenerateFullName() string {
	parts := []string{}

	if c.Prefix != "" {
		parts = append(parts, c.Prefix)
	}
	if c.GivenName != "" {
		parts = append(parts, c.GivenName)
	}
	if c.MiddleName != "" {
		parts = append(parts, c.MiddleName)
	}
	if c.FamilyName != "" {
		parts = append(parts, c.FamilyName)
	}
	if c.Suffix != "" {
		parts = append(parts, c.Suffix)
	}

	if len(parts) == 0 {
		if c.Nickname != "" {
			return c.Nickname
		}
		return "Unnamed Contact"
	}

	var fullName strings.Builder
	for i, part := range parts {
		if i > 0 {
			fullName.WriteString(" ")
		}
		fullName.WriteString(part)
	}
	return fullName.String()
}

// HasAnniversary returns true if either the full date or the partial components are set
func (c *Contact) HasAnniversary() bool {
	hasFullDate := c.Anniversary != nil
	hasPartial := c.AnniversaryMonth != nil && c.AnniversaryDay != nil
	return hasFullDate || hasPartial
}

// HasAvatar returns true if both an avatar and mimetype are set
func (c *Contact) HasAvatar() bool {
	return c.AvatarBase64 != "" && c.AvatarMimeType != ""
}

// HasAvatar returns true if both an avatar and mimetype are set
func (c *Contact) HasFullBirthday() bool {
	return c.Birthday != nil
}

// ContactWithRelationships extends Contact with relationship information
type ContactWithRelationships struct {
	*Contact
	Relationships []Relationship `json:"relationships,omitempty"`
}
