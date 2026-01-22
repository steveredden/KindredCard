package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/steveredden/KindredCard/internal/immich"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/models"
)

// ===== IMMICH WEB PAGES =====

// ImmichLinkPage renders the Immich sync status page
func (h *Handler) ImmichLinkPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if Immich is configured
	immichURL := os.Getenv("IMMICH_URL")
	immichToken := os.Getenv("IMMICH_KEY")

	if immichURL == "" || immichToken == "" {
		h.renderTemplate(w, r, "util_immich_link.html", map[string]interface{}{
			"User":       user,
			"Title":      "Immich Person Linker",
			"Configured": false,
			"Error":      "Immich integration not configured.\nPlease set IMMICH_URL and IMMICH_KEY in your .env file.",
		})
		return
	}

	// Create Immich client
	client := immich.NewClient(immichURL, immichToken)

	// Test connection
	if err := client.TestConnection(); err != nil {
		logger.Error("[IMMICH] Connection test failed: %v", err)
		return
	}

	syncService := immich.NewSyncService(client, h.db, user.ID)

	potential, _ := syncService.GetPotentialMatches()
	existing, _ := syncService.GetAllLinkedContacts()

	h.renderTemplate(w, r, "util_immich_link.html", map[string]interface{}{
		"User":       user,
		"Title":      "Immich Person Linker",
		"Configured": true,
		"Items":      potential,
		"Count":      len(potential),
		"Existing":   existing,
	})
}

// ImmichLinkPage renders the Immich sync status page
func (h *Handler) ImmichManagementPage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if Immich is configured
	immichURL := os.Getenv("IMMICH_URL")
	immichToken := os.Getenv("IMMICH_KEY")

	if immichURL == "" || immichToken == "" {
		h.renderTemplate(w, r, "util_immich_link.html", map[string]interface{}{
			"User":       user,
			"Title":      "Immich Link Management & Sync",
			"Configured": false,
			"Error":      "Immich integration not configured.\nPlease set IMMICH_URL and IMMICH_KEY in your .env file.",
		})
		return
	}

	// Create Immich client
	client := immich.NewClient(immichURL, immichToken)

	// Test connection
	if err := client.TestConnection(); err != nil {
		logger.Error("[IMMICH] Connection test failed: %v", err)
		return
	}

	syncService := immich.NewSyncService(client, h.db, user.ID)

	existing, _ := syncService.GetAllLinkedContacts()

	h.renderTemplate(w, r, "util_immich_manage.html", map[string]interface{}{
		"User":       user,
		"Title":      "Immich Link Management & Sync",
		"Configured": true,
		"Items":      existing,
		"Count":      len(existing),
		"Base":       immichURL,
	})
}

// ===== IMMICH API ENDPOINTS =====

// PostImmichLinkAPI handles the actual linking of a contact to an Immich ID
func (h *Handler) PostImmichLinkAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ContactID int    `json:"contact_id"`
		PersonID  string `json:"person_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s/people/%s", os.Getenv("IMMICH_URL"), req.PersonID)

	// Create the URL record with type ["immich"]
	newURL := models.URL{
		ContactID: req.ContactID,
		URL:       url,
		Type:      []string{"immich"},
	}

	_, err := h.db.CreateContactURL(user.ID, newURL)
	if err != nil {
		http.Error(w, "Failed to save link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetImmichThumbnailProxy handles proxying authenticated image requests to Immich
func (h *Handler) GetImmichThumbnailProxy(w http.ResponseWriter, r *http.Request) {
	// Get Contact ID from URL
	personID := mux.Vars(r)["personID"]
	if personID == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	client := immich.NewClient(os.Getenv("IMMICH_URL"), os.Getenv("IMMICH_KEY"))
	thumbData, err := client.GetPersonThumbnail(personID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg") // Immich thumbnails are usually jpegs
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(thumbData)
}
