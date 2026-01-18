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
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// Main Index (index.html) Page!
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {

	// on the index page cache the user context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	contacts, err := h.db.GetAllContacts(user.ID, false)
	if err != nil {
		http.Error(w, "Error loading contacts", http.StatusInternalServerError)
		return
	}

	// Get counters
	totalCount, _ := h.db.GetContactCount(user.ID)
	upcomingEventCount, _ := h.db.GetUpcomingEventsCount(user.ID, 7)
	recentlyEditedCount, _ := h.db.GetRecentlyEditedCountByDays(user.ID, 7)

	h.renderTemplate(w, "index.html", map[string]interface{}{
		"User":               user,
		"Contacts":           contacts,
		"TotalContacts":      totalCount,
		"UpcomingEventCount": upcomingEventCount,
		"RecentCount":        recentlyEditedCount,
		"Title":              "Contacts",
		"ActivePage":         "contacts",
	})

	token, _ := middleware.GetTokenFromCurrentSession(r)
	h.db.UpdateSessionActivity(token)
}

// ShowContact displays a single contact
func (h *Handler) ShowContact(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	contact, err := h.db.GetContactByID(user.ID, id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	// format dates
	birthdayView := getBirthdayView(*contact)
	anniversaryView := getAnniversaryView(*contact)
	otherDatesView := getOtherDatesView(contact.OtherDates)

	// for new relationship creation
	allContacts, _ := h.db.GetAllContactsAbbrv(user.ID, false)
	relationshipTypes, _ := h.db.GetRelationshipTypes()

	h.renderTemplate(w, "contact_detail.html", map[string]interface{}{
		"AllContacts":       allContacts,
		"RelationshipTypes": relationshipTypes,
		"Contact":           contact,
		"Birthday":          birthdayView,
		"Anniversary":       anniversaryView,
		"OtherDates":        otherDatesView,
		"User":              user,
		"Title":             contact.FullName,
		"ActivePage":        "contacts",
	})

	token, _ := middleware.GetTokenFromCurrentSession(r)
	h.db.UpdateSessionActivity(token)
}

// PartialDateView represents a date that may or may not have a year
type PartialDateView struct {
	Has       bool
	Month     int
	Day       int
	MonthName string
}

// OtherDateView represents an other_date with PartialDateView
type OtherDateView struct {
	ID        int
	EventName string
	Date      PartialDateView
	Year      int
}

func getBirthdayView(contact models.Contact) PartialDateView {

	var birthday PartialDateView

	switch {
	case contact.Birthday != nil:
		birthday = PartialDateView{
			Has:       true,
			Month:     int(contact.Birthday.Month()),
			Day:       contact.Birthday.Day(),
			MonthName: contact.Birthday.Format("January"),
		}

	case contact.BirthdayMonth != nil && contact.BirthdayDay != nil:
		t := time.Date(2000, time.Month(*contact.BirthdayMonth), *contact.BirthdayDay, 0, 0, 0, 0, time.UTC)
		birthday = PartialDateView{
			Has:       true,
			Month:     *contact.BirthdayMonth,
			Day:       *contact.BirthdayDay,
			MonthName: t.Format("January"),
		}
	}

	return birthday
}

func getAnniversaryView(contact models.Contact) PartialDateView {

	var anniversary PartialDateView

	switch {
	case contact.Anniversary != nil:
		anniversary = PartialDateView{
			Has:       true,
			Month:     int(contact.Anniversary.Month()),
			Day:       contact.Anniversary.Day(),
			MonthName: contact.Anniversary.Format("January"),
		}

	case contact.AnniversaryMonth != nil && contact.AnniversaryDay != nil:
		t := time.Date(2000, time.Month(*contact.AnniversaryMonth), *contact.AnniversaryDay, 0, 0, 0, 0, time.UTC)
		anniversary = PartialDateView{
			Has:       true,
			Month:     *contact.AnniversaryMonth,
			Day:       *contact.AnniversaryDay,
			MonthName: t.Format("January"),
		}
	}

	return anniversary
}

// getOtherDatesView converts OtherDates slice to OtherDateView slice
func getOtherDatesView(otherDates []models.OtherDate) []OtherDateView {
	if len(otherDates) == 0 {
		return []OtherDateView{}
	}

	views := make([]OtherDateView, 0, len(otherDates))

	for _, od := range otherDates {
		view := OtherDateView{
			ID:        od.ID,
			EventName: od.EventName,
			Date:      getPartialDateView(od),
		}

		// Add year if available from full date
		if od.EventDate != nil {
			view.Year = od.EventDate.Year()
		}

		views = append(views, view)
	}

	return views
}

// getPartialDateView converts an OtherDate to PartialDateView
func getPartialDateView(otherDate models.OtherDate) PartialDateView {
	var dateView PartialDateView

	switch {
	case otherDate.EventDate != nil:
		// Full date with year
		dateView = PartialDateView{
			Has:       true,
			Month:     int(otherDate.EventDate.Month()),
			Day:       otherDate.EventDate.Day(),
			MonthName: otherDate.EventDate.Format("Jan"),
		}
	case otherDate.EventDateMonth != nil && otherDate.EventDateDay != nil:
		// Partial date (month/day only)
		t := time.Date(2000, time.Month(*otherDate.EventDateMonth), *otherDate.EventDateDay, 0, 0, 0, 0, time.UTC)
		dateView = PartialDateView{
			Has:       true,
			Month:     *otherDate.EventDateMonth,
			Day:       *otherDate.EventDateDay,
			MonthName: t.Format("Jan"),
		}
	}

	return dateView
}

// HandlePatchContact godoc
//
//	@Summary		Partially update a contact
//	@Description	Update specific fields of a contact using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Contact ID"
//	@Param			contact	body		models.ContactJSONPatch	true	"Contact fields to update"
//	@Success		200		{object}	models.Contact			"Updated contact"
//	@Failure		400		{object}	map[string]string		"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		404		{object}	map[string]string		"Contact not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{id} [patch]
func (h *Handler) PatchContactAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	contactID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse PATCH request
	var patch models.ContactJSONPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply patch
	updated, err := h.db.PatchContact(user.ID, contactID, &patch)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	// Return updated record
	json.NewEncoder(w).Encode(updated)
}
