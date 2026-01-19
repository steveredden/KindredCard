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

	"github.com/steveredden/KindredCard/internal/middleware"
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

	h.renderTemplate(w, "util_gender_assign.html", map[string]interface{}{
		"User":     user,
		"Contacts": contacts,
	})

	token, _ := middleware.GetTokenFromCurrentSession(r)
	h.db.UpdateSessionActivity(token)
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

	h.renderTemplate(w, "util_phone_formatter.html", map[string]interface{}{
		"User":     user,
		"Contacts": contacts,
	})

	token, _ := middleware.GetTokenFromCurrentSession(r)
	h.db.UpdateSessionActivity(token)
}
