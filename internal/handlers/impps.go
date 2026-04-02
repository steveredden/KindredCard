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

// NewIMPPAPI godoc
//
//	@Summary		Creates a website / IMPP associated with a contact
//	@Description	Associates a IMPP with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Contact ID"
//	@Param			impp	body		models.IMPP			true	"IMPP fields"
//	@Success		200		{object}	map[string]string	"created: newID"
//	@Failure		400		{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		404		{object}	map[string]string	"Contact not found"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/impp [post]
func (h *Handler) NewIMPPAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from IMPP
	contactID, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.IMPP

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	input.ContactID = contactID

	newID, err := h.db.CreateContactIMPP(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("created: %d", newID)})
}

// UpdateIMPPAPI godoc
//
//	@Summary		Update a impp number
//	@Description	Update specific fields of a impp using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			iid		path		int						true	"IMPP ID"
//	@Param			contact	body		models.IMPPJSONPatch	true	"IMPP fields to update"
//	@Success		200		{object}	[]models.IMPP			"Updated impp"
//	@Failure		400		{object}	map[string]string		"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		404		{object}	map[string]string		"Contact not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/impps/{iid} [patch]
func (h *Handler) UpdateIMPPAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get IMPP ID from IMPP
	imppID, err := strconv.Atoi(mux.Vars(r)["iid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.IMPPJSONPatch

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	input.ID = imppID

	updated, err := h.db.UpdateContactIMPP(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	// Return updated record
	json.NewEncoder(w).Encode(updated)
}

// DeleteIMPPAPI godoc
//
//	@Summary		Creates a website / IMPP associated with a contact
//	@Description	Associates a IMPP with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			cid	path		int					true	"Contact ID"
//	@Param			iid	path		int					true	"IMPP ID"
//	@Success		200	{object}	map[string]string	"deleted"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/impp/{iid} [delete]
func (h *Handler) DeleteIMPPAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Contact ID from IMPP
	contactID, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "Invalid Contact ID", http.StatusBadRequest)
		return
	}

	// Get IMPP ID from IMPP
	imppID, err := strconv.Atoi(mux.Vars(r)["iid"])
	if err != nil {
		http.Error(w, "Invalid IMPP ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactIMPP(user.ID, contactID, imppID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
