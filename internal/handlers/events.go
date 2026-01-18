/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

// ShowEvents displays the events page
func (h *Handler) ShowEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	// Fetch past events (last 7 days)
	pastEvents, err := h.db.GetRecentPastEventsByDays(user.ID, 7)
	if err != nil {
		http.Error(w, "Failed to fetch past events", http.StatusInternalServerError)
		return
	}
	pastEventCount := len(pastEvents)

	// Insert "Today" marker only if we have past events
	if pastEventCount > 0 {
		nowUTC := time.Now().UTC()
		todayMidnight := time.Date(
			nowUTC.Year(),
			nowUTC.Month(),
			nowUTC.Day(),
			0, 0, 0, 0,
			time.UTC, // <-- Forces the location to +0000 Z
		)

		todayEvent := models.UpcomingEvent{
			ContactID:       -42,
			FullName:        "TODAY_MARKER",
			EventType:       "today-marker",
			EventLabel:      "    ---------- TODAY ----------",
			EventDate:       &todayMidnight,
			ThisYearDate:    todayMidnight,
			DaysUntil:       0,
			TimeDescription: "Today",
		}

		// Insert the fake event
		pastEvents = append(pastEvents, todayEvent)
	}

	// Fetch upcoming events (next 7 days)
	upcomingEvents, err := h.db.GetUpcomingEventsByDays(user.ID, 7)
	if err != nil {
		http.Error(w, "Failed to fetch upcoming events", http.StatusInternalServerError)
		return
	}
	upcomingEventCount := len(upcomingEvents)

	// Fetch events for next 3 months (90 days)
	events, err := h.db.GetUpcomingEventsByMonths(user.ID, 3)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	// Count today's events
	todayCount := countTodayEvents(events)

	// Combine the timeline
	combinedEvents := make([]models.UpcomingEvent, 0)
	combinedEvents = append(combinedEvents, pastEvents...)
	combinedEvents = append(combinedEvents, events...)

	// Group events by month
	monthlyGroups := groupEventsByMonth(combinedEvents)

	h.renderTemplate(w, "events.html", map[string]interface{}{
		"User":          user,
		"Title":         "Events",
		"ActivePage":    "events",
		"MonthlyEvents": monthlyGroups,
		"PastCount":     pastEventCount,
		"TodayCount":    todayCount,
		"UpcomingCount": upcomingEventCount,
	})

	token, _ := middleware.GetTokenFromCurrentSession(r)
	h.db.UpdateSessionActivity(token)
}

// GetUpcomingEventsAPI godoc
//
//	@Summary		Get upcoming events
//	@Description	Get birthdays, anniversaries, and other important dates coming up for all contacts
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			days	query		int						false	"Number of days to look ahead"	default(30)	minimum(1)	maximum(365)
//	@Param			type	query		string					false	"Filter by event type"			enums(birthday,anniversary,other)
//	@Success		200		{array}		models.UpcomingEvent	"List of upcoming events"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/events/upcoming [get]
func (h *Handler) GetUpcomingEventsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	// Get query parameters
	timeframe := r.URL.Query().Get("timeframe") // "days" or "months"
	value := r.URL.Query().Get("value")         // "2" or "3", etc.

	// Parse value
	val, err := strconv.Atoi(value)
	if err != nil {
		val = 2 // Default
	}

	var events []models.UpcomingEvent

	// Get events based on timeframe
	if timeframe == "months" {
		events, err = h.db.GetUpcomingEventsByMonths(user.ID, val)
	} else {
		// Default to days
		events, err = h.db.GetUpcomingEventsByDays(user.ID, val)
	}

	if err != nil {
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetTodaysEventsHandler handles GET /api/v1/events/today
func (h *Handler) GetTodaysEventsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	events, err := h.db.GetTodaysEvents(user.ID)
	if err != nil {
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetUpcomingEventsCountHandler handles GET /api/v1/events/count
func (h *Handler) GetUpcomingEventsCountAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	// Default to 3 days
	days := 3
	if d := r.URL.Query().Get("days"); d != "" {
		if val, err := strconv.Atoi(d); err == nil {
			days = val
		}
	}

	count, err := h.db.GetUpcomingEventsCount(user.ID, days)
	if err != nil {
		http.Error(w, "Failed to get count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"count": count,
	})
}

// Helper functions

// countTodayEvents counts events happening today (days_until == 0)
func countTodayEvents(events []models.UpcomingEvent) int {
	count := 0
	for _, event := range events {
		if event.DaysUntil == 0 {
			count++
		}
	}
	return count
}

// groupEventsByMonth groups events by month and formats them
func groupEventsByMonth(events []models.UpcomingEvent) []models.MonthlyEventGroup {
	// Map to hold events by month key (YYYY-MM)
	monthMap := make(map[string]*models.MonthlyEventGroup)

	// Track unique month keys for ordering
	var monthKeys []string

	for _, event := range events {
		// Use the non-pointer ThisYearDate, which is guaranteed to be non-nil
		// and correctly calculated by the database for the timeline view.
		eventTime := event.ThisYearDate

		// Create month key (e.g., "2026-01")
		monthKey := eventTime.Format("2006-01")

		// Initialize month group if it doesn't exist
		if _, exists := monthMap[monthKey]; !exists {
			monthMap[monthKey] = &models.MonthlyEventGroup{
				MonthName: eventTime.Format("Jan 2006"),
				Year:      eventTime.Year(),
				Month:     eventTime.Month(),
				Events:    []models.TimelineEventItem{},
			}
			monthKeys = append(monthKeys, monthKey)
		}

		// Create timeline item with formatted description
		item := models.TimelineEventItem{
			ContactID:        event.ContactID,
			FullName:         event.FullName,
			DayFormatted:     eventTime.Format("Jan 02"),
			EventDate:        eventTime,
			EventDescription: formatEventDescription(event),
		}

		monthMap[monthKey].Events = append(monthMap[monthKey].Events, item)
	}

	// Sort month keys chronologically
	sort.Strings(monthKeys)

	// Build final sorted list
	var result []models.MonthlyEventGroup
	for _, key := range monthKeys {
		group := monthMap[key]

		// Sort events within each month by day
		sort.Slice(group.Events, func(i, j int) bool {
			eventI := group.Events[i]
			eventJ := group.Events[j]

			// 1. Primary Sort: Sort by date
			if !eventI.EventDate.Equal(eventJ.EventDate) {
				return eventI.EventDate.Before(eventJ.EventDate)
			}

			// 2. Secondary Sort (Tiebreaker): If dates are equal, force the marker (-1) to the front
			return eventI.ContactID < eventJ.ContactID
		})

		result = append(result, *group)
	}

	return result
}

// formatEventDescription formats the event description based on type and available data
func formatEventDescription(event models.UpcomingEvent) string {
	switch event.EventType {
	case "today-marker":
		return event.EventLabel
	case "birthday":
		if event.AgeOrYears != nil && *event.AgeOrYears > 0 {
			return utils.Ordinal(*event.AgeOrYears) + " birthday"
		}
		return "birthday"

	case "anniversary":
		if event.AgeOrYears != nil && *event.AgeOrYears > 0 {
			return utils.Ordinal(*event.AgeOrYears) + " wedding anniversary"
		}
		return "wedding anniversary"

	default:
		// For custom date events
		label := event.EventType
		if label == "" {
			label = "event"
		}

		if event.AgeOrYears != nil && *event.AgeOrYears > 0 {
			return utils.Ordinal(*event.AgeOrYears) + " anniversary of " + label
		}
		return "anniversary of " + label
	}
}
