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

// NewOrganizationAPI godoc
//
//	@Summary		Creates a website / URL associated with a contact
//	@Description	Associates a URL with a contact using HTTP POST
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int					true	"Contact ID"
//	@Param			url	body		models.Organization	true	"Organization fields"
//	@Success		200	{object}	map[string]string	"created: newID"
//	@Failure		400	{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Contact not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/organizations [post]
func (h *Handler) NewOrganizationAPI(w http.ResponseWriter, r *http.Request) {
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

	var input models.Organization

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Error("[HANDLER] Could not parse input: %v", err)
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	input.ContactID = contactID

	newID, err := h.db.CreateContactOrganization(user.ID, input)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("created: %d", newID)})
}

// UpdateOrganizationAPI godoc
//
//	@Summary		Update a organization number
//	@Description	Update specific fields of a organization using HTTP PATCH. Only provided fields will be updated.
//	@Tags			contacts
//	@Accept			json
//	@Produce		json
//	@Param			oid		path		int								true	"Organization ID"
//	@Param			contact	body		models.OrganizationJSONPatch	true	"Organization fields to update"
//	@Success		200		{object}	[]models.Organization			"Updated organization"
//	@Failure		400		{object}	map[string]string				"Invalid request body or contact ID"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Failure		404		{object}	map[string]string				"Contact not found"
//	@Failure		500		{object}	map[string]string				"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/organizations/{oid} [patch]
func (h *Handler) UpdateOrganizationAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Organization ID from URL
	organizationID, err := strconv.Atoi(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input models.OrganizationJSONPatch

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", 400)
		return
	}

	input.ID = organizationID

	updated, err := h.db.UpdateContactOrganization(user.ID, input)
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
//	@Param			cid				path		int					true	"Contact ID"
//	@Param			organizationid	path		int					true	"URL ID"
//	@Success		200				{object}	map[string]string	"deleted"
//	@Failure		400				{object}	map[string]string	"Invalid request body or contact ID"
//	@Failure		401				{object}	map[string]string	"Unauthorized"
//	@Failure		404				{object}	map[string]string	"Contact not found"
//	@Failure		500				{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/contacts/{cid}/organizations/{organizationid} [delete]
func (h *Handler) DeleteOrganizationAPI(w http.ResponseWriter, r *http.Request) {
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
	organizationID, err := strconv.Atoi(mux.Vars(r)["aid"])
	if err != nil {
		http.Error(w, "Invalid URL ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactOrganization(user.ID, contactID, organizationID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
