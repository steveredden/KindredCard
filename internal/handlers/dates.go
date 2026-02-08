package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// NewOtherDateAPI godoc
//
//	@Summary		Creates an Other Date associated with a contact
//	@Description	Associates an Other Date with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid			path		int					true	"Contact ID"
//	@Param			otherDate	body		models.OtherDate	true	"Other Date fields"
//	@Success		200			{object}	map[string]string	"created: newID"
//	@Failure		400			{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401			{object}	map[string]string	"Unauthorized"
//	@Failure		404			{object}	map[string]string	"Contact not found"
//	@Failure		500			{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/other-dates [post]
func (h *Handler) NewOtherDateAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from Other Date
	contactID, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.ContactDateJSONPatch

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	input.ContactID = contactID

	newID, err := h.db.CreateContactOtherDate(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("created: %d", newID)})
}

// UpdateOtherDateAPI godoc
//
//	@Summary		Update an other_date for a contact
//	@Description	Update specific fields of a phone using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			oid		path		int							true	"Other Date ID"
//	@Param			contact	body		models.ContactDateJSONPatch	true	"Other Date fields to update"
//	@Success		200		{object}	[]models.OtherDate			"Updated other date"
//	@Failure		400		{object}	map[string]string			"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Failure		404		{object}	map[string]string			"Contact not found"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/other-dates/{oid} [patch]
func (h *Handler) UpdateOtherDateAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from URL
	otherDateID, err := strconv.Atoi(mux.Vars(r)["oid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	input := models.ContactDateJSONPatch{
		EventID:  &otherDateID,
		DateType: "other",
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	err = h.db.UpdateContactOtherDate(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// DeleteOtherDateAPI godoc
//
//	@Summary		Removes an Other Date associated with a contact
//	@Description	Removes an Other Date with a contact using HTTP DELETE
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid	path		int					true	"Contact ID"
//	@Param			oid	path		int					true	"Other Date ID"
//	@Success		200	{object}	map[string]string	"deleted"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/other-dates/{oid} [delete]
func (h *Handler) DeleteOtherDateAPI(w http.ResponseWriter, r *http.Request) {
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

	// Get URL ID from URL
	otherDateID, err := strconv.Atoi(mux.Vars(r)["oid"])
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactOtherDate(user.ID, contactID, otherDateID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
