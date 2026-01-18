/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package utils

import (
	"encoding/json"
	"strconv"
	"time"
)

// NullableDate is a custom type for handling dates that can be null
type NullableDate struct {
	Time  time.Time
	Valid bool
}

// UnmarshalJSON implements custom JSON unmarshaling for dates
func (nd *NullableDate) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		nd.Valid = false
		return nil
	}

	// Handle empty string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str == "" {
		nd.Valid = false
		return nil
	}

	// Try parsing as date-only format first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		// Try parsing as RFC3339 (datetime)
		t, err = time.Parse(time.RFC3339, str)
		if err != nil {
			nd.Valid = false
			return nil
		}
	}

	nd.Time = t
	nd.Valid = true
	return nil
}

// MarshalJSON implements custom JSON marshaling for dates
func (nd NullableDate) MarshalJSON() ([]byte, error) {
	if !nd.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nd.Time.Format("2006-01-02"))
}

// ToTimePtr converts NullableDate to *time.Time for database storage
func (nd *NullableDate) ToTimePtr() *time.Time {
	if !nd.Valid {
		return nil
	}
	return &nd.Time
}

// FromTimePtr creates a NullableDate from *time.Time
func FromTimePtr(t *time.Time) NullableDate {
	if t == nil {
		return NullableDate{Valid: false}
	}
	return NullableDate{Time: *t, Valid: true}
}

// ParseDateString parses a date string and returns a *time.Time
// Handles empty strings, "YYYY-MM-DD" format, and null values
func ParseDateString(s string) (*time.Time, error) {
	if s == "" || s == "null" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// FormatDate formats a *time.Time as YYYY-MM-DD or empty string if nil
func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateLong formats a *time.Time as "January 2, 2006" or empty string if nil
func FormatDateLong(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("January 2, 2006")
}

// FormatPartialDate formats month and day pointers as "January 2" or empty string
func FormatPartialDate(month, day *int) string {
	if month == nil || day == nil {
		return ""
	}

	months := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	if *month < 1 || *month > 12 || *day < 1 || *day > 31 {
		return ""
	}

	return months[*month] + " " + strconv.Itoa(*day)
}

// getMonth returns the month from a time.Time pointer
func GetMonth(t *time.Time) int {
	if t == nil {
		return 0
	}
	return int(t.Month())
}

// getDay returns the day from a time.Time pointer
func GetDay(t *time.Time) int {
	if t == nil {
		return 0
	}
	return t.Day()
}

// getYear returns the year from a time.Time pointer
// Returns 0 if nil (which won't be displayed in the input)
func GetYear(t *time.Time) int {
	if t == nil {
		return 0
	}
	// If year is 1 (our sentinel for partial dates), return 0
	if t.Year() == 1 {
		return 0
	}
	return t.Year()
}

func FormatBirthday(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("January 2, 2006")
}

func FormatBirthdayShort(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Jan 2")
}

// formatBirthdayMedium formats a date as "March 15, 1976" or "March 15" for partial dates
func FormatBirthdayMedium(t *time.Time) string {
	if t == nil {
		return ""
	}

	// Partial date (year = 1)
	if t.Year() == 1 {
		return t.Format("January 2")
	}

	// Full date
	return t.Format("January 2, 2006")
}
