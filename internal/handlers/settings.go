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

	"github.com/gorilla/mux"

	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/notifications"
)

// Contacts Setting Page

func (h *Handler) DeleteAllContactsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete all contacts
	err := h.db.DeleteAllContacts(user.ID)
	if err != nil {
		http.Error(w, "Failed to delete contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "All contacts deleted successfully",
	})
}

// FindDuplicatesHandler scans for duplicate contacts
func (h *Handler) FindDuplicatesAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Find duplicates
	duplicates, err := h.db.FindDuplicateContacts(user.ID)
	if err != nil {
		http.Error(w, "Failed to find duplicates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"duplicates": duplicates,
		"count":      len(duplicates),
	})
}

// Notification Settings Handlers

func (h *Handler) CreateNotificationSettingAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	var req models.NotificationSetting

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	notifier, err := h.db.CreateNotificationSetting(user.ID, &req)
	if err != nil {
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notifier)
}

func (h *Handler) GetNotificationSettingAPI(w http.ResponseWriter, r *http.Request) {
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

	notification, err := h.db.GetNotificationSettingByID(user.ID, id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

func (h *Handler) ListNotificationSettingsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	notifications, err := h.db.GetAllUserNotificationSettings(user.ID)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

func (h *Handler) UpdateNotificationSettingAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	var req models.NotificationSetting

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.db.UpdateNotificationSetting(user.ID, &req); err != nil {
		http.Error(w, "Error updating contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) DeleteNotificationSettingAPI(w http.ResponseWriter, r *http.Request) {
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

	if err := h.db.DeleteNotificationSetting(user.ID, id); err != nil {
		http.Error(w, "Error updating contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) TestNotificationAPI(w http.ResponseWriter, r *http.Request) {
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

	// Get user's notification settings
	settings, err := h.db.GetNotificationSettingByID(user.ID, id)
	if err != nil || settings.WebhookURL == "" {
		http.Error(w, "No webhook URL configured", http.StatusBadRequest)
		return
	}

	// Send test notification
	err = notifications.SendTestNotification(settings.WebhookURL, h.baseURL)
	if err != nil {
		http.Error(w, "Failed to send test notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ========================================
// SECURITY HANDLERS
// ========================================

// DeleteAccountHandler deletes the user account
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete user and all associated data
	err := h.db.DeleteUser(user.ID)
	if err != nil {
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	h.Logout(w, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account deleted successfully",
	})
}

// DeleteUserSession deletes a user session
func (h *Handler) DeleteUserSessionAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	sessionID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	err = h.db.RevokeSessionByID(user.ID, sessionID)
	if err != nil {
		http.Error(w, "Failed to delete session Token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Session deleted successfully",
	})
}

// DeleteUserSession deletes a user session
func (h *Handler) DeleteAllOtherUserSessionsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, ok := middleware.GetTokenFromCurrentSession(r)
	if !ok {
		http.Error(w, "Failed to find user sessions", http.StatusInternalServerError)
		return
	}

	allSessions, err := h.db.GetUserSessions(user.ID)
	if err != nil {
		http.Error(w, "Failed to find sessions", http.StatusInternalServerError)
		return
	}

	for _, session := range allSessions {
		if session.Token == token {
			continue
		}
		_ = h.db.RevokeSessionByID(user.ID, session.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sessions deleted successfully",
	})
}
