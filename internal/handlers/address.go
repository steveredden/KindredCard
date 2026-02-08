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

// NewAddressAPI godoc
//
//	@Summary		Creates a website / URL associated with a contact
//	@Description	Associates a URL with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int					true	"Contact ID"
//	@Param			url	body		models.Address		true	"Address fields"
//	@Success		200	{object}	map[string]string	"created: newID"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/addresses [post]
func (h *Handler) NewAddressAPI(w http.ResponseWriter, r *http.Request) {
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

	var input models.Address

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("[HANDLER] Could not parse input: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	input.ContactID = contactID

	newID, err := h.db.CreateContactAddress(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("created: %d", newID)})
}

// UpdateAddressAPI godoc
//
//	@Summary		Update a address number
//	@Description	Update specific fields of a address using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			aid		path		int						true	"Address ID"
//	@Param			contact	body		models.AddressJSONPatch	true	"Address fields to update"
//	@Success		200		{object}	[]models.Address		"Updated address"
//	@Failure		400		{object}	map[string]string		"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		404		{object}	map[string]string		"Contact not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/addresses/{aid} [patch]
func (h *Handler) UpdateAddressAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Address ID from URL
	addressID, err := strconv.Atoi(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.AddressJSONPatch

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	input.ID = addressID

	updated, err := h.db.UpdateContactAddress(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	// Return updated record
	json.NewEncoder(w).Encode(updated)
}

// DeleteURLAPI godoc
//
//	@Summary		Creates a website / URL associated with a contact
//	@Description	Associates a URL with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid	path		int					true	"Contact ID"
//	@Param			aid	path		int					true	"URL ID"
//	@Success		200	{object}	map[string]string	"deleted"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/addresses/{a} [delete]
func (h *Handler) DeleteAddressAPI(w http.ResponseWriter, r *http.Request) {
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
	addressID, err := strconv.Atoi(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactAddress(user.ID, contactID, addressID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
