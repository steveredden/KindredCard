package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// NewCustomLabelAPI godoc
//
//	@Summary		Get upcoming events
//	@Description	Get birthdays, anniversaries, and other important dates coming up for all contacts
//	@Tags			labels
//	@Accept			json
//	@Produce		json
//	@Param			label	body		models.ContactLabelJSONPost	true	"label fields"
//	@Success		200		{array}		map[string]int				"id of label"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/settings/labels [post]
func (h *Handler) NewCustomLabelAPI(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r)
	if !ok {
		return
	}

	var input models.ContactLabelJSONPost

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	label, err := h.db.NewLabel(input.Name, input.Category)
	if err != nil {
		http.Error(w, "Failed to create label", http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"id": label})
}

// DeleteEmailAPI godoc
//
//	@Summary		Removes a custom contact label
//	@Description	Removes a custom contact label using HTTP DELETE
//	@Tags			labels
//	@Accept			json
//	@Produce		json
//	@Param			lid	path		int					true	"Label ID"
//	@Success		200	{object}	map[string]string	"deleted"
//	@Failure		400	{object}	map[string]string	"Invalid request body or Label ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Label not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Security		ApiTokenAuth
//	@Router			/api/v1/settings/labels/{lid} [delete]
func (h *Handler) DeleteCustomLabelAPI(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Label ID from URL
	labelID, err := strconv.Atoi(mux.Vars(r)["lid"])
	if err != nil {
		http.Error(w, "Invalid Contact ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteContactLabel(labelID)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
