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
	"time"
)

// ContactJSON is used for JSON marshaling/unmarshaling with proper date handling
type ContactJSON struct {
	ID                 int                 `json:"id"`
	UID                string              `json:"uid"`
	FullName           string              `json:"full_name"`
	GivenName          string              `json:"given_name"`
	FamilyName         string              `json:"family_name"`
	MiddleName         string              `json:"middle_name"`
	Prefix             string              `json:"prefix"`
	Suffix             string              `json:"suffix"`
	Nickname           string              `json:"nickname"`
	Gender             string              `json:"gender,omitempty"`
	Birthday           string              `json:"birthday,omitempty"` // String for flexible parsing
	BirthdayMonth      *int                `json:"birthday_month,omitempty"`
	BirthdayDay        *int                `json:"birthday_day,omitempty"`
	Anniversary        string              `json:"anniversary,omitempty"` // String for flexible parsing
	AnniversaryMonth   *int                `json:"anniversary_month,omitempty"`
	AnniversaryDay     *int                `json:"anniversary_day,omitempty"`
	Notes              string              `json:"notes"`
	AvatarBase64       string              `json:"avatar_base64,omitempty"`
	AvatarMimeType     string              `json:"avatar_mime_type,omitempty"`
	ExcludeFromSync    bool                `json:"exclude_from_sync"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	ETag               string              `json:"etag"`
	Emails             []Email             `json:"emails,omitempty"`
	Phones             []Phone             `json:"phones,omitempty"`
	Addresses          []Address           `json:"addresses,omitempty"`
	Organizations      []Organization      `json:"organizations,omitempty"`
	URLs               []URL               `json:"urls,omitempty"`
	Relationships      []Relationship      `json:"relationships,omitempty"`
	OtherRelationships []OtherRelationship `json:"other_relationships,omitempty"`
	OtherDates         []OtherDateJSON     `json:"other_dates,omitempty"`
}

// OtherDateJSON is used for JSON marshaling/unmarshaling of other dates
type OtherDateJSON struct {
	ID             int    `json:"id,omitempty"`
	EventName      string `json:"event_name"`
	EventDate      string `json:"event_date,omitempty"` // Full date string "2020-01-15"
	EventDateMonth *int   `json:"event_date_month,omitempty"`
	EventDateDay   *int   `json:"event_date_day,omitempty"`
}

// ContactJSON is used for JSON marshaling/unmarshaling with proper date handling
type ContactJSONPatch struct {
	GivenName       *string `json:"given_name" example:"John"`
	FamilyName      *string `json:"family_name" example:"Doe"`
	MiddleName      *string `json:"middle_name" example:"J"`
	Prefix          *string `json:"prefix" example:"Dr."`
	Suffix          *string `json:"suffix" example:"Jr."`
	Nickname        *string `json:"nickname" example:"Johnny"`
	Gender          *string `json:"gender,omitempty" example:"M" enums:"M,F,O"`
	Notes           *string `json:"notes" example:"VIP"`
	AvatarBase64    *string `json:"avatar_base64,omitempty"`
	AvatarMimeType  *string `json:"avatar_mime_type,omitempty"`
	ExcludeFromSync *bool   `json:"exclude_from_sync" example:"false"`
}

type ContactDateJSONPatch struct {
	ContactID int        `json:"id" example:"123"`
	DateType  string     `json:"date_type" example:"anniversary"` // "anniversary", "birthday", or "other"
	EventID   *int       `json:"other_date_id" example:"456"`
	Date      *time.Time `json:"date" example:"2026-04-30"`
	DateMonth *int       `json:"date_month" example:"4"`
	DateDay   *int       `json:"date_day" example:"30"`
}

type PhoneJSONPatch struct {
	ID    *int    `json:"id" example:"1"`
	Phone *string `json:"phone" example:"(555)122-4121"`
}

// OtherDateJSON is used for JSON marshaling/unmarshaling of other dates
type OtherDateJSONPatch struct {
	EventName      *string `json:"event_name"`
	EventDate      *string `json:"event_date,omitempty"` // Full date string "2020-01-15"
	EventDateMonth *int    `json:"event_date_month,omitempty"`
	EventDateDay   *int    `json:"event_date_day,omitempty"`
}

// ToContact converts ContactJSON to Contact
func (cj *ContactJSON) ToContact() (*Contact, error) {
	contact := &Contact{
		ID:                 cj.ID,
		UID:                cj.UID,
		FullName:           cj.FullName,
		GivenName:          cj.GivenName,
		FamilyName:         cj.FamilyName,
		MiddleName:         cj.MiddleName,
		Prefix:             cj.Prefix,
		Suffix:             cj.Suffix,
		Nickname:           cj.Nickname,
		Gender:             cj.Gender,
		BirthdayMonth:      cj.BirthdayMonth,
		BirthdayDay:        cj.BirthdayDay,
		AnniversaryMonth:   cj.AnniversaryMonth,
		AnniversaryDay:     cj.AnniversaryDay,
		Notes:              cj.Notes,
		AvatarBase64:       cj.AvatarBase64,
		AvatarMimeType:     cj.AvatarMimeType,
		ExcludeFromSync:    cj.ExcludeFromSync,
		CreatedAt:          cj.CreatedAt,
		UpdatedAt:          cj.UpdatedAt,
		ETag:               cj.ETag,
		Emails:             cj.Emails,
		Phones:             cj.Phones,
		Addresses:          cj.Addresses,
		Organizations:      cj.Organizations,
		URLs:               cj.URLs,
		Relationships:      cj.Relationships,
		OtherRelationships: cj.OtherRelationships,
	}

	// Parse birthday string if provided
	if cj.Birthday != "" {
		t, err := time.Parse("2006-01-02", cj.Birthday)
		if err == nil {
			contact.Birthday = &t
		}
	}

	// Parse anniversary string if provided
	if cj.Anniversary != "" {
		t, err := time.Parse("2006-01-02", cj.Anniversary)
		if err == nil {
			contact.Anniversary = &t
		}
	}

	// Convert OtherDateJSON to OtherDate
	if len(cj.OtherDates) > 0 {
		contact.OtherDates = make([]OtherDate, 0, len(cj.OtherDates))
		for _, odj := range cj.OtherDates {
			od := OtherDate{
				ID:             odj.ID,
				EventName:      odj.EventName,
				EventDateMonth: odj.EventDateMonth,
				EventDateDay:   odj.EventDateDay,
			}

			// Parse event_date string if provided
			if odj.EventDate != "" {
				t, err := time.Parse("2006-01-02", odj.EventDate)
				if err == nil {
					od.EventDate = &t
				}
			}

			contact.OtherDates = append(contact.OtherDates, od)
		}
	}

	return contact, nil
}

// FromContact converts Contact to ContactJSON
func FromContact(contact *Contact) *ContactJSON {
	cj := &ContactJSON{
		ID:                 contact.ID,
		UID:                contact.UID,
		FullName:           contact.FullName,
		GivenName:          contact.GivenName,
		FamilyName:         contact.FamilyName,
		MiddleName:         contact.MiddleName,
		Prefix:             contact.Prefix,
		Suffix:             contact.Suffix,
		Nickname:           contact.Nickname,
		Gender:             contact.Gender,
		BirthdayMonth:      contact.BirthdayMonth,
		BirthdayDay:        contact.BirthdayDay,
		AnniversaryMonth:   contact.AnniversaryMonth,
		AnniversaryDay:     contact.AnniversaryDay,
		Notes:              contact.Notes,
		AvatarBase64:       contact.AvatarBase64,
		AvatarMimeType:     contact.AvatarMimeType,
		ExcludeFromSync:    contact.ExcludeFromSync,
		CreatedAt:          contact.CreatedAt,
		UpdatedAt:          contact.UpdatedAt,
		ETag:               contact.ETag,
		Emails:             contact.Emails,
		Phones:             contact.Phones,
		Addresses:          contact.Addresses,
		Organizations:      contact.Organizations,
		URLs:               contact.URLs,
		Relationships:      contact.Relationships,
		OtherRelationships: contact.OtherRelationships,
	}

	// Format birthday as string if provided
	if contact.Birthday != nil {
		cj.Birthday = contact.Birthday.Format("2006-01-02")
	}

	// Format anniversary as string if provided
	if contact.Anniversary != nil {
		cj.Anniversary = contact.Anniversary.Format("2006-01-02")
	}

	// Convert OtherDate to OtherDateJSON
	if len(contact.OtherDates) > 0 {
		cj.OtherDates = make([]OtherDateJSON, 0, len(contact.OtherDates))
		for _, od := range contact.OtherDates {
			odj := OtherDateJSON{
				ID:             od.ID,
				EventName:      od.EventName,
				EventDateMonth: od.EventDateMonth,
				EventDateDay:   od.EventDateDay,
			}

			// Format event_date as string if provided
			if od.EventDate != nil {
				odj.EventDate = od.EventDate.Format("2006-01-02")
			}

			cj.OtherDates = append(cj.OtherDates, odj)
		}
	}

	return cj
}

// Helper method to add to ContactJSONPatch model
func (p *ContactJSONPatch) HasUpdates() bool {
	return p.GivenName != nil ||
		p.FamilyName != nil ||
		p.MiddleName != nil ||
		p.Prefix != nil ||
		p.Suffix != nil ||
		p.Nickname != nil ||
		p.Gender != nil ||
		p.Notes != nil ||
		p.AvatarBase64 != nil ||
		p.AvatarMimeType != nil ||
		p.ExcludeFromSync != nil
}
