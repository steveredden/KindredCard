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
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Helper: initial
func Initial(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) == 0 {
		return "?"
	}
	return strings.ToUpper(string(r[0]))
}

func GenderFullString(s string) string {
	switch s {
	case "M":
		return "Male"
	case "F":
		return "Female"
	case "N":
		return "" //prefer not to say
	case "O":
		return "Other"
	}

	return ""
}

// Helper: monthName - conver month number to name
func MonthName(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return time.Month(month).String()
}

// Helper: MonthNameShort - convert month number to short name (Jan, Feb, ...)
func MonthNameShort(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	return time.Month(month).String()[:3]
}

// Helper: add
func Add(a, b int) int {
	return a + b
}

// ParseIntPtr parses a string to *int, returns nil if empty
func ParseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

func IntPtr(val int) *int {
	return &val
}

// DerefInt dereferences an int pointer, returns 0 if nil
func DerefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// NullString helpers
func ScanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// NullInt helpers
func ScanNullInt(ni sql.NullInt64) *int {
	if ni.Valid {
		val := int(ni.Int64)
		return &val
	}
	return nil
}

func ToNullInt(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*i), Valid: true}
}

// NullTime helpers
func ScanNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func ToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func TruncateWebhook(url string) string {
	if len(url) > 50 {
		return url[:47] + "..."
	}
	return url
}

// iterate creates a slice of integers from 1 to n
// Usage: {{range $day := iterate 31}}
func Iterate(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i + 1
	}
	return result
}

// getPrimaryOrFirst returns the primary item or first item from a slice
// Works with both Email and Phone slices
func GetPrimaryOrFirst(items interface{}, fieldName string) string {
	switch v := items.(type) {
	case []struct {
		Email     string
		Type      string
		IsPrimary bool
	}:
		// Find primary email
		for _, item := range v {
			if item.IsPrimary {
				return item.Email
			}
		}
		// Return first if no primary
		if len(v) > 0 {
			return v[0].Email
		}

	case []struct {
		Phone     string
		Type      string
		IsPrimary bool
	}:
		// Find primary phone
		for _, item := range v {
			if item.IsPrimary {
				return item.Phone
			}
		}
		// Return first if no primary
		if len(v) > 0 {
			return v[0].Phone
		}
	}

	return ""
}

// calculateYears calculates years from an event date
func CalculateYears(event *time.Time) int {
	if event == nil {
		return 0
	}
	// Don't calculate for partial dates (year = 1)
	if event.Year() == 1 {
		return 0
	}

	now := time.Now()
	years := now.Year() - event.Year()

	// Adjust if event hasn't occurred this year yet
	if now.Month() < event.Month() ||
		(now.Month() == event.Month() && now.Day() < event.Day()) {
		years--
	}

	return years
}

// Template helper functions

// formatEventType formats the event type with emoji
func FormatEventType(eventType string) string {
	switch eventType {
	case "birthday":
		return "🎂 Birthday"
	case "anniversary":
		return "💍 Anniversary"
	default:
		return "📅 " + eventType
	}
}

// timelineColor returns the color class for timeline items
func TimelineColor(daysUntil int) string {
	if daysUntil < 0 {
		return "warning"
	} else if daysUntil == 0 {
		return "primary"
	}
	return "success"
}

// timelineBg returns the background color class for timeline items
func TimelineBg(daysUntil int) string {
	if daysUntil < 0 {
		return "bg-warning/10"
	} else if daysUntil == 0 {
		return "bg-primary/10"
	}
	return "bg-success/10"
}

// Helper: Format date time for display
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return t.Local().Format("Jan 02, 2006 at 3:04 PM")
}

// Helper: Format date time pointer for display
func FormatDateTimePtr(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return FormatDateTime(*t)
}

func HasType(types []string, target string) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

// Ordinal converts a number to its Ordinal form (1 -> "1st", 2 -> "2nd", etc.)
func Ordinal(n int) string {
	if n <= 0 {
		return ""
	}

	suffix := "th"
	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}

	return fmt.Sprintf("%d%s", n, suffix)
}

func ExtractIDFromImmichURL(url string) string {
	parts := strings.Split(strings.TrimRight(url, "/"), "/")
	personID := parts[len(parts)-1]
	return personID
}
