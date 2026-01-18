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

// NotificationSettings represents user notification preferences
type NotificationSetting struct {
	ID                   int        `json:"id"`
	Name                 string     `json:"name"`
	UserID               int        `json:"user_id"`
	WebhookURL           string     `json:"webhook_url"`
	DaysLookAhead        int        `json:"days_look_ahead"`
	NotificationTime     string     `json:"notification_time"` // HH:MM format
	IncludeBirthdays     bool       `json:"include_birthdays"`
	IncludeAnniversaries bool       `json:"include_anniversaries"`
	IncludeEventDates    bool       `json:"include_event_dates"`
	EventRegex           string     `json:"other_event_regex"`
	Enabled              bool       `json:"enabled"`
	LastSentAt           *time.Time `json:"last_sent_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	CreatedAt            time.Time  `json:"created_at"`
}

type ContactStats struct {
	TotalContacts  int
	AddedThisMonth int
	WithBirthdays  int
}

type DuplicateGroup struct {
	Contact1ID   int    `json:"contact1_id"`
	Contact1Name string `json:"contact1_name"`
	Contact2ID   int    `json:"contact2_id"`
	Contact2Name string `json:"contact2_name"`
	MatchType    string `json:"match_type"` // "name" or "email"
}
