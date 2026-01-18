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
	"fmt"
	"time"
)

// UpcomingEvent represents an event (birthday, anniversary, or other date) with timing information
type UpcomingEvent struct {
	ContactID       int        `json:"contact_id"`
	FullName        string     `json:"full_name"`
	EventType       string     `json:"event_type"`  // "birthday", "anniversary", something else
	EventLabel      string     `json:"event_label"` // custom label
	EventDate       *time.Time `json:"event_date"`  // The actual date of the event
	ThisYearDate    time.Time  `json:"this_year_date"`
	DaysUntil       int        `json:"days_until"`       // Negative = past, 0 = today, positive = future
	AgeOrYears      *int       `json:"age_or_years"`     // Age for birthdays, years for anniversaries (nullable)
	TimeDescription string     `json:"time_description"` // "Yesterday", "Today", "Tomorrow", "3 days ago", "in 5 days"
}

// MonthlyEventGroup represents events grouped by month
type MonthlyEventGroup struct {
	MonthName string              // "Jan 2026"
	Year      int                 // 2026
	Month     time.Month          // 1 (January)
	Events    []TimelineEventItem // Sorted events for this month
}

// TimelineEventItem represents a single event in the timeline
type TimelineEventItem struct {
	ContactID        int
	FullName         string
	DayFormatted     string // "Jan 01", "Feb 24"
	EventDate        time.Time
	EventDescription string // "14th birthday", "1st anniversary of engagement", "birthday"
}

// IsToday returns true if the event is happening today
func (e *UpcomingEvent) IsToday() bool {
	return e.DaysUntil == 0
}

// IsPast returns true if the event has already occurred
func (e *UpcomingEvent) IsPast() bool {
	return e.DaysUntil < 0
}

// IsFuture returns true if the event hasn't occurred yet
func (e *UpcomingEvent) IsFuture() bool {
	return e.DaysUntil > 0
}

// GetColorClass returns the appropriate color class for UI rendering
func (e *UpcomingEvent) GetColorClass() string {
	if e.IsPast() {
		return "warning"
	} else if e.IsToday() {
		return "primary"
	}
	return "success"
}

// GetBgClass returns the appropriate background color class for UI rendering
func (e *UpcomingEvent) GetBgClass() string {
	if e.IsPast() {
		return "bg-warning/10"
	} else if e.IsToday() {
		return "bg-primary/10"
	}
	return "bg-success/10"
}

// GetBadgeClass returns the badge style class
func (e *UpcomingEvent) GetBadgeClass() string {
	if e.IsPast() {
		return "badge-warning"
	} else if e.IsToday() {
		return "badge-primary"
	}
	return "badge-success"
}

// GetAlertClass returns the alert style class
func (e *UpcomingEvent) GetAlertClass() string {
	if e.IsPast() {
		return "alert-warning"
	} else if e.IsToday() {
		return "alert-primary"
	}
	return "alert-success"
}

// GetFormattedEventType returns the event type with emoji
func (e *UpcomingEvent) GetFormattedEventType() string {
	switch e.EventType {
	case "birthday":
		return "🎂 Birthday"
	case "anniversary":
		return "💍 Anniversary"
	case "other":
		if e.EventLabel != "" {
			return "📅 " + e.EventLabel
		}
		return "📅 Event"
	default:
		return "📅 " + e.EventType
	}
}

// GetAgeOrYearsText returns formatted text for age/years
func (e *UpcomingEvent) GetAgeOrYearsText() string {
	if e.AgeOrYears == nil {
		return ""
	}

	value := *e.AgeOrYears

	switch e.EventType {
	case "birthday":
		if e.IsPast() {
			return fmt.Sprintf("Turned %d", value)
		}
		return fmt.Sprintf("Turning %d", value)
	case "anniversary":
		if e.IsPast() {
			return fmt.Sprintf("%d years", value)
		}
		return fmt.Sprintf("%d years", value)
	default:
		return ""
	}
}

// GetActionButtonText returns the appropriate button text
func (e *UpcomingEvent) GetActionButtonText() string {
	if e.IsPast() {
		return "Send belated wishes"
	} else if e.IsToday() {
		return "Send wishes now!"
	}
	return "View contact"
}
