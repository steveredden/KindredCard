package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// NewEmailAPI godoc
//
//	@Summary		Creates an email associated with a contact
//	@Description	Associates an email with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid	path		int					true	"Contact ID"
//	@Param			url	body		models.Email		true	"Email fields"
//	@Success		200	{object}	map[string]string	"created: newID"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/emails [post]
func (h *Handler) NewEmailAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from URL
	contactID, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.Email

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("[HANDLER] Could not parse input: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	input.ContactID = contactID

	newID, err := h.db.CreateContactEmail(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("created: %d", newID)})
}

// UpdateEmailAPI godoc
//
//	@Summary		Update a email number
//	@Description	Update specific fields of a email using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			eid		path		int						true	"Email ID"
//	@Param			contact	body		models.EmailJSONPatch	true	"Email fields to update"
//	@Success		200		{object}	[]models.Email			"Updated email"
//	@Failure		400		{object}	map[string]string		"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		404		{object}	map[string]string		"Contact not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/emails/{eid} [patch]
func (h *Handler) UpdateEmailAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Email ID from URL
	emailID, err := strconv.Atoi(mux.Vars(r)["eid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.EmailJSONPatch

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	input.ID = emailID

	updated, err := h.db.UpdateContactEmail(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	// Return updated record
	json.NewEncoder(w).Encode(updated)
}

// DeleteEmailAPI godoc
//
//	@Summary		Removes an email associated with a contact
//	@Description	Associates an email with a contact using HTTP DELETE
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid	path		int					true	"Contact ID"
//	@Param			eid	path		int					true	"Email ID"
//	@Success		200	{object}	map[string]string	"deleted"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/emails/{eid} [delete]
func (h *Handler) DeleteEmailAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from URL
	contactID, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "Invalid Contact ID", http.StatusBadRequest)
		return
	}

	// Get email ID from URL
	emailID, err := strconv.Atoi(mux.Vars(r)["eid"])
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactEmail(user.ID, contactID, emailID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
