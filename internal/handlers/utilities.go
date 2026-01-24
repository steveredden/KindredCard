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
	"net/http"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (h *Handler) GenderAssignmentPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contacts, err := h.db.GetContactsMissingGender(user.ID)
	if err != nil {
		http.Error(w, "Database error", 500)
		return
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of contacts:")
		utils.Dump(contacts)
	}

	h.renderTemplate(w, r, "util_gender_assign.html", map[string]interface{}{
		"Title": "Gender Assigner",
		"User":  user,
		"Items": contacts,
		"Count": len(contacts),
	})
}

func (h *Handler) PhoneFormatterPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	contacts, err := h.db.GetContactsWithPhones(user.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of contacts:")
		utils.Dump(contacts)
	}

	h.renderTemplate(w, r, "util_phone_formatter.html", map[string]interface{}{
		"Title": "Phone Formatter",
		"User":  user,
		"Items": contacts,
		"Count": len(contacts),
	})
}

func (h *Handler) RelationshipAssignmentPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	suggestions, err := h.db.GetRelationshipSuggestions(user.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of suggestions:")
		utils.Dump(suggestions)
	}

	h.renderTemplate(w, r, "util_relationship_assign.html", map[string]interface{}{
		"Title": "Relationship Assignment",
		"User":  user,
		"Items": suggestions,
		"Count": len(suggestions),
	})
}

func (h *Handler) AnniversaryProposalPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	suggestions, err := h.db.GetAnniversarySuggestions(user.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of suggestions:")
		utils.Dump(suggestions)
	}

	h.renderTemplate(w, r, "util_anniversary_proposer.html", map[string]interface{}{
		"Title": "Anniversary Proposal",
		"User":  user,
		"Items": suggestions,
		"Count": len(suggestions),
	})
}

func (h *Handler) AddressProposalPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	suggestions, err := h.db.GetAddressSuggestions(user.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[HANDLER] Dump of suggestions:")
		utils.Dump(suggestions)
	}

	h.renderTemplate(w, r, "util_address_proposer.html", map[string]interface{}{
		"Title": "Address Proposal",
		"User":  user,
		"Items": suggestions,
		"Count": len(suggestions),
	})
}
