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

// Contact represents a person in the CRM
type Contact struct {
	ID                 int                 `json:"id" example:"1"`
	UID                string              `json:"uid" example:"42"`
	FullName           string              `json:"full_name" example:"Dr. John Frank Doe III"` // Computed from name parts
	GivenName          string              `json:"given_name" example:"John"`
	FamilyName         string              `json:"family_name" example:"Doe"`
	MiddleName         string              `json:"middle_name" example:"Frank"`
	Prefix             string              `json:"prefix" example:"Dr."`
	Suffix             string              `json:"suffix" example:"III"`
	Nickname           string              `json:"nickname" example:"broheim"`
	Gender             string              `json:"gender,omitempty" example:"M"` // M, F, O, N, U
	Birthday           *time.Time          `json:"birthday,omitempty" example:"1990-12-15T00:00:00Z"`
	BirthdayMonth      *int                `json:"birthday_month,omitempty" example:"12"` // 1-12, for partial dates
	BirthdayDay        *int                `json:"birthday_day,omitempty" example:"15"`   // 1-31, for partial dates
	Anniversary        *time.Time          `json:"anniversary,omitempty" example:"2022-01-03T00:00:00Z"`
	AnniversaryMonth   *int                `json:"anniversary_month,omitempty" example:"1"` // 1-12, for partial dates
	AnniversaryDay     *int                `json:"anniversary_day,omitempty" example:"3"`   // 1-31, for partial dates
	Notes              string              `json:"notes" example:"Met at work conference 2023"`
	AvatarBase64       string              `json:"avatar_base64,omitempty"`
	AvatarMimeType     string              `json:"avatar_mime_type,omitempty"`
	ExcludeFromSync    bool                `json:"exclude_from_sync"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	ETag               string              `json:"etag"`
	LastModifiedToken  int                 `json:"last_modified_token"`
	UserID             int                 `json:"user_id"`
	Emails             []Email             `json:"emails,omitempty"`
	Phones             []Phone             `json:"phones,omitempty"`
	Addresses          []Address           `json:"addresses,omitempty"`
	Organizations      []Organization      `json:"organizations,omitempty"`
	OtherDates         []OtherDate         `json:"other_dates"`
	URLs               []URL               `json:"urls,omitempty"`
	Relationships      []Relationship      `json:"relationships,omitempty"`
	OtherRelationships []OtherRelationship `json:"other_relationships,omitempty"`
	DeletedAt          *time.Time
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

	fullName := ""
	for i, part := range parts {
		if i > 0 {
			fullName += " "
		}
		fullName += part
	}
	return fullName
}

// Email represents an email address
type Email struct {
	ID        int      `json:"id"`
	ContactID int      `json:"contact_id"`
	Email     string   `json:"email"`
	Type      []string `json:"type"` // home, work, other
	IsPrimary bool     `json:"is_primary"`
}

// Phone represents a phone number
type Phone struct {
	ID        int      `json:"id"`
	ContactID int      `json:"contact_id"`
	Phone     string   `json:"phone"`
	Type      []string `json:"type"` // home, work, mobile, fax, other
	IsPrimary bool     `json:"is_primary"`
}

// Address represents a physical address
type Address struct {
	ID         int      `json:"id"`
	ContactID  int      `json:"contact_id"`
	Street     string   `json:"street"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	PostalCode string   `json:"postal_code"`
	Country    string   `json:"country"`
	Type       []string `json:"type"` // home, work, other
	IsPrimary  bool     `json:"is_primary"`
}

// Organization represents a company/organization affiliation
type Organization struct {
	ID         int    `json:"id"`
	ContactID  int    `json:"contact_id"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	Department string `json:"department"`
	IsPrimary  bool   `json:"is_primary"`
}

// URL represents a website or online profile
type URL struct {
	ID        int      `json:"id"`
	ContactID int      `json:"contact_id"`
	URL       string   `json:"url"`
	Type      []string `json:"type"` // website, social, other
	IsPrimary bool     `json:"is_primary"`
}

// Tag represents a label/category for contacts
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ContactWithRelationships extends Contact with relationship information
type ContactWithRelationships struct {
	*Contact
	Relationships []Relationship `json:"relationships,omitempty"`
}

// Other Dated Events
type OtherDate struct {
	ID             int        `json:"id"`
	ContactID      int        `json:"contact_id"`
	EventName      string     `json:"event_name"`
	EventDate      *time.Time `json:"event_date,omitempty"`
	EventDateMonth *int       `json:"event_date_month,omitempty"` // 1-12, for partial dates
	EventDateDay   *int       `json:"event_date_day,omitempty"`   // 1-31, for partial dates
}
