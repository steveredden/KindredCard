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
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/emersion/go-vcard"
	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/converter"
	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

type Handler struct {
	db             *db.Database
	templates      *template.Template
	user           *models.User
	baseURL        string
	releaseVersion string
}

func NewHandler(database *db.Database, templatesPath string, baseURL string, releaseVersion string) (*Handler, error) {
	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"appVersion":           func() string { return releaseVersion },
			"add":                  utils.Add,
			"deref":                utils.DerefInt,
			"formatDate":           utils.FormatDate,
			"formatDateLong":       utils.FormatDateLong,
			"formatPartialDate":    utils.FormatPartialDate,
			"formatBirthday":       utils.FormatBirthday,
			"formatBirthdayShort":  utils.FormatBirthdayShort,
			"initial":              utils.Initial,
			"monthName":            utils.MonthName,
			"monthNameShort":       utils.MonthNameShort,
			"truncateWebhook":      utils.TruncateWebhook,
			"iterate":              utils.Iterate,
			"getMonth":             utils.GetMonth,
			"getDay":               utils.GetDay,
			"getYear":              utils.GetYear,
			"formatBirthdayMedium": utils.FormatBirthdayMedium,
			"calculateYears":       utils.CalculateYears,
			"genderFullString":     utils.GenderFullString,
			"formatEventType":      utils.FormatEventType,
			"timelineColor":        utils.TimelineColor,
			"timelineBg":           utils.TimelineBg,
			"formatDateTimePtr":    utils.FormatDateTimePtr,
			"formatDateTime":       utils.FormatDateTime,
			"hasType":              utils.HasType,
			"isCustom":             utils.IsCustom,
		}).
		ParseGlob(templatesPath + "/*.html")

	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.ParseGlob(templatesPath + "/components/*.html")
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.ParseGlob(templatesPath + "/pages/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:             database,
		templates:      tmpl,
		baseURL:        baseURL,
		releaseVersion: releaseVersion,
	}, nil
}

// renderTemplate renders a template by combining base layout with specific page
func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, pageName string, data any) error {
	token, _ := middleware.GetTokenFromCurrentSession(r)
	if token != "" {
		h.db.UpdateSessionActivity(token)
	}

	tmpl, err := h.templates.Clone()
	if err != nil {
		return err
	}

	_, err = tmpl.ParseFiles(
		"web/templates/pages/" + pageName,
	)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// Web Interface Handlers

// ShowSettings displays the settings page
func (h *Handler) ShowSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	// Get contact statistics
	stats, err := h.db.GetContactStats(user.ID)
	if err != nil {
		stats = &models.ContactStats{} // Default empty stats
	}

	userNotifications, err := h.db.GetAllUserNotificationSettings(user.ID)
	if err != nil {
		userNotifications = []models.NotificationSetting{}
	}

	// userPreferences, _ := h.db.GetUserPreferences(user.ID)

	tokens, err := h.db.GetAPITokensByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch API tokens", http.StatusInternalServerError)
		return
	}

	// Get current session token
	currentToken, _ := middleware.GetTokenFromCurrentSession(r)

	// premptively update session activity, since we're soon hitting the page that renders said last_activity
	if currentToken != "" {
		h.db.UpdateSessionActivity(currentToken)
	}

	// Get all sessions for this user
	sessions, err := h.db.GetUserSessions(user.ID)
	if err != nil {
		http.Error(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}

	// Mark current session
	for i := range sessions {
		if sessions[i].Token == currentToken {
			sessions[i].IsCurrent = true
		}
	}

	h.renderTemplate(w, r, "settings.html", map[string]interface{}{
		"User":                 user,
		"Title":                "Settings",
		"ActivePage":           "settings",
		"Stats":                stats,
		"APITokens":            tokens,
		"Sessions":             sessions,
		"NotificationSettings": userNotifications,
	})
}

// API Handlers

// ListContactsAPI returns all contacts as JSON
// ListContactsAPI godoc
//
//	@Summary		Lists all contacts
//	@Description	Get all contacts for the authenticated user
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Security		SessionAuth
//	@Success		200	{array}		models.Contact
//	@Failure		401	{object}	models.ErrorResponse
//	@Failure		500	{object}	models.ErrorResponse
//	@Router			/contacts [get]
func (h *Handler) ListContactsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	contacts, err := h.db.GetAllContacts(user.ID, false)
	if err != nil {
		http.Error(w, "Error loading contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

// GetContactAPI godoc
//
//	@Summary		Get a single contact
//	@Description	Retrieve detailed information about a specific contact by ID
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int					true	"Contact ID"	minimum(1)
//	@Success		200	{object}	models.Contact		"Contact details"
//	@Failure		400	{object}	map[string]string	"Invalid contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{id} [get]
func (h *Handler) GetContactAPI(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

// CreateContactAPI godoc
//
//	@Summary		Create a new contact
//	@Description	Create a new contact with the provided information
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			contact	body		models.Contact		true	"Contact information"
//	@Success		201		{object}	models.Contact		"Created contact"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		409		{object}	map[string]string	"Contact already exists"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts [post]
func (h *Handler) CreateContactAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}
	var contact models.Contact

	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contact.FullName = contact.GenerateFullName()

	if err := h.db.CreateContact(user.ID, &contact); err != nil {
		http.Error(w, "Error creating contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

// UpdateContactAPI godoc
//
//	@Summary		Update a contact (full replacement)
//	@Description	Replace all fields of an existing contact. Use PATCH for partial updates.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Contact ID"	minimum(1)
//	@Param			contact	body		models.ContactJSON	true	"Updated contact information"
//	@Success		200		{object}	models.Contact		"Updated contact"
//	@Failure		400		{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		404		{object}	map[string]string	"Contact not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{id} [put]
func (h *Handler) UpdateContactAPI(w http.ResponseWriter, r *http.Request) {
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

	var contactJSON models.ContactJSON
	if err := json.NewDecoder(r.Body).Decode(&contactJSON); err != nil {
		logger.Error("[HANDLER] Error decoding contact data: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contact, err := contactJSON.ToContact()
	if err != nil {
		http.Error(w, "Invalid contact data", http.StatusBadRequest)
		return
	}

	// Quick fix for #6 - if saved from the GUI then don't delete and insert relationships
	queryParams := r.URL.Query().Get("source")
	if queryParams != "" && queryParams == "GUI" {
		contact.Metadata = "skip relationships"
	}

	contact.ID = id
	contact.FullName = contact.GenerateFullName()

	// If partial dates are provided, clear full dates
	if contact.BirthdayMonth != nil && contact.BirthdayDay != nil {
		contact.Birthday = nil
	}
	if contact.AnniversaryMonth != nil && contact.AnniversaryDay != nil {
		contact.Anniversary = nil
	}

	if err := h.db.UpdateContact(user.ID, contact); err != nil {
		http.Error(w, "Error updating contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

// DeleteContactAPI godoc
//
//	@Summary		Delete a contact
//	@Description	Soft delete a contact by ID. The contact will be marked as deleted but not permanently removed.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id			path	int		true	"Contact ID"						minimum(1)
//	@Param			permanent	query	bool	false	"Permanently delete (hard delete)"	default(false)
//	@Success		204			"Contact deleted successfully"
//	@Failure		400			{object}	map[string]string	"Invalid contact ID"
//	@Failure		401			{object}	map[string]string	"Unauthorized"
//	@Failure		404			{object}	map[string]string	"Contact not found"
//	@Failure		500			{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{id} [delete]
func (h *Handler) DeleteContactAPI(w http.ResponseWriter, r *http.Request) {
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

	if err := h.db.DeleteContact(user.ID, id); err != nil {
		http.Error(w, "Error deleting contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchContactsAPI searches contacts
func (h *Handler) SearchContactsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}

	contacts, err := h.db.SearchContacts(user.ID, query)
	if err != nil {
		http.Error(w, "Error searching contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

// GetRelationshipTypesAPI returns all available relationship types
func (h *Handler) GetRelationshipTypesAPI(w http.ResponseWriter, r *http.Request) {
	types, err := h.db.GetRelationshipTypes()
	if err != nil {
		http.Error(w, "Error loading relationship types", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

// AddRelationshipAPI godoc
//
//	@Summary		Add relationship to contact
//	@Description	Create a relationship between two contacts (e.g., spouse, child, parent)
//	@Tags			relationships
//	@Accept			json
//	@Produce		json
//	@Param			id				path		int					true	"Contact ID"	minimum(1)
//	@Param			relationship	body		models.Relationship	true	"Relationship details"
//	@Success		201				{object}	models.Relationship	"Created relationship"
//	@Failure		400				{object}	map[string]string	"Invalid request body"
//	@Failure		401				{object}	map[string]string	"Unauthorized"
//	@Failure		404				{object}	map[string]string	"Contact not found"
//	@Failure		409				{object}	map[string]string	"Relationship already exists"
//	@Failure		500				{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{id}/relationships [post]
func (h *Handler) AddRelationshipAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req struct {
		RelatedContactID   int `json:"related_contact_id"`
		RelationshipTypeID int `json:"relationship_type_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.db.AddRelationship(user.ID, contactID, req.RelatedContactID, req.RelationshipTypeID); err != nil {
		http.Error(w, "Error adding relationship", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}

// RemoveRelationshipAPI deletes a relationship
func (h *Handler) RemoveRelationshipAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	relIDStr := vars["rel_id"]

	relID, err := strconv.Atoi(relIDStr)
	if err != nil {
		http.Error(w, "Invalid relationship ID", http.StatusBadRequest)
		return
	}

	if err := h.db.RemoveRelationship(user.ID, relID); err != nil {
		http.Error(w, "Error removing relationship", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRelationshipAPI deletes a relationship
func (h *Handler) RemoveOtherRelationshipAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	relIDStr := vars["rel_id"]

	relID, err := strconv.Atoi(relIDStr)
	if err != nil {
		http.Error(w, "Invalid other relationship ID", http.StatusBadRequest)
		return
	}

	if err := h.db.RemoveOtherRelationship(user.ID, relID); err != nil {
		http.Error(w, "Error removing relationship", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadAvatarAPI handles avatar uploads
func (h *Handler) UploadAvatarAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Avatar string `json:"avatar"` // base64 encoded
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Avatar) == 0 {
		http.Error(w, "Avatar data required", http.StatusBadRequest)
		return
	}

	// Detect MIME type
	mimeType := "image/jpeg"
	if len(req.Avatar) > 10 {
		switch req.Avatar[0] {
		case '/':
			mimeType = "image/jpeg"
		case 'i':
			mimeType = "image/png"
		case 'R':
			mimeType = "image/gif"
		}
	}

	err = h.db.UpdateAvatar(user.ID, contactID, req.Avatar, mimeType)
	if err != nil {
		http.Error(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DeleteAvatarAPI handles avatar removals
func (h *Handler) DeleteAvatarAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	contactIDStr := vars["id"]

	contactID, err := strconv.Atoi(contactIDStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteAvatar(user.ID, contactID)
	if err != nil {
		http.Error(w, "Failed to delete avatar", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ExportContactVCardAPI exports a single contact as vCard
func (h *Handler) ExportContactVCardAPI(w http.ResponseWriter, r *http.Request) {
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
		logger.Error("[HANDLER] Error retriving contact: %v", err)
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	// Convert to vCard
	card := converter.ContactToVCard(contact, false)

	var buf bytes.Buffer
	encoder := vcard.NewEncoder(&buf)
	if err := encoder.Encode(card); err != nil {
		logger.Error("[HANDLER] Error encoding vCard: %v", err)
		http.Error(w, "Error encoding vCard", http.StatusInternalServerError)
		return
	}

	// Set headers for download
	filename := contact.FullName
	if filename == "" {
		filename = "contact"
	}
	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".vcf\"")
	w.Write(buf.Bytes())
}

// ExportAllVCardsAPI godoc
//
//	@Summary		Export all contacts as vCard
//	@Description	Download all contacts in vCard (.vcf) format
//	@Tags			export
//	@Produce		text/vcard
//	@Success		200	{file}		file				"vCard file download"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/export/vcard [get]
func (h *Handler) ExportAllVCardsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	contacts, err := h.db.GetAllContacts(user.ID, false) // Get all contacts
	if err != nil {
		http.Error(w, "Error loading contacts", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	encoder := vcard.NewEncoder(&buf)

	for _, contact := range contacts {
		card := converter.ContactToVCard(contact, false)
		if err := encoder.Encode(card); err != nil {
			continue
		}
	}

	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"kindredcard-contacts.vcf\"")
	w.Write(buf.Bytes())
}

// ExportAllJSONAPI exports all contacts as JSON
func (h *Handler) ExportAllJSONAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	contacts, err := h.db.GetAllContacts(user.ID, false) // Get all contacts
	if err != nil {
		http.Error(w, "Error loading contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"kindredcard-contacts.json\"")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contacts": contacts,
		"exported": len(contacts),
	})
}

// ImportVCards imports contacts from uploaded vCard file
func (h *Handler) ImportVCardsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		logger.Error("[HANDLER] Error parsing multipartform: %v", err)
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("vcard")
	if err != nil {
		logger.Error("[HANDLER] Error retreiving form file: %v", err)
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	content, _ := io.ReadAll(file)

	// Pass 0: Decode all vCards into a slice immediately
	var cards []vcard.Card
	decoder := vcard.NewDecoder(bytes.NewReader(content))
	for {
		card, err := decoder.Decode()
		if err == io.EOF {
			break
		}
		if err == nil {
			cards = append(cards, card)
		}
	}

	// Pass 1: Create "Shells"
	// We only care about UID and FullName here to satisfy FKs for relationships
	for _, card := range cards {
		contact, err := converter.VCardToContactShell(card)
		if err != nil {
			continue
		}

		// if input doesnt have a UID, we created one -> and now need to set it for loop 2
		if card.Get(vcard.FieldUID) == nil {
			card.SetValue(vcard.FieldUID, contact.UID)
		}

		// Create the contact if it doesn't exist
		// h.db.CreateContact should handle "ON CONFLICT (user_id, uid) DO NOTHING"
		_ = h.db.CreateContact(user.ID, contact)
	}

	// Fetch current state for relationship matching
	allContacts, _ := h.db.GetAllContactsAbbrv(user.ID, false)
	allRelTypes, _ := h.db.GetRelationshipTypes()

	// Create a quick lookup map: UID -> Internal ID
	uidToID := make(map[string]int)
	for _, c := range allContacts {
		uidToID[c.UID] = c.ID
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of uidToID:")
		utils.Dump(uidToID)
	}

	// Pass 2: Full Update
	// Now converter.VCardToContact can find the related contacts in allContacts
	imported := 0
	for _, card := range cards {
		contact, err := converter.VCardToContact(card, allContacts, allRelTypes)
		if err != nil {
			logger.Debug("[HANDLER] Error converting vCard to Contact: %v", err)
			continue
		}

		//Populate the contact.ID based on what's been created or already exists
		logger.Trace("UID: %s", contact.UID)
		if id, ok := uidToID[contact.UID]; ok {
			contact.ID = id
			if err := h.db.UpdateContact(user.ID, contact); err == nil {
				imported++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  imported,
		"status": "success",
	})
}

// User preferences API
func (h *Handler) UpdatePreferencesAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	var userPref models.User

	if err := json.NewDecoder(r.Body).Decode(&userPref); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userPref.ID = user.ID

	// Update preferences
	err := h.db.UpdateUserPreferences(userPref)
	if err != nil {
		// Handle specific DB errors if needed, otherwise send generic server error
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
