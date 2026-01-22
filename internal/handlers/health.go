package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/steveredden/KindredCard/internal/immich"
	"github.com/steveredden/KindredCard/internal/logger"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`    // "ok" or "error"
	Timestamp string            `json:"timestamp"` // Current time
	Version   string            `json:"version"`   // App version
	Checks    map[string]string `json:"checks"`    // Individual component checks
}

// HandleHealth godoc
//
//	@Summary		Health check endpoint
//	@Description	Returns the health status of the application and its dependencies
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	HealthResponse	"Service is healthy"
//	@Failure		503	{object}	HealthResponse	"Service is unhealthy"
//	@Router			/health [get]
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	logger.Debug("[HEALTH] Health check requested")

	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   h.releaseVersion,
		Checks:    make(map[string]string),
	}

	// Check database connection
	if err := h.db.Ping(); err != nil {
		logger.Error("[HEALTH] Database health check failed: %v", err)
		response.Status = "error"
		response.Checks["database"] = "unhealthy: " + err.Error()
	} else {
		response.Checks["database"] = "healthy"
	}

	// Check smtp if used (non-fatal)
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpHost != "" && smtpPort != "" {
		smtpAddr := net.JoinHostPort(smtpHost, smtpPort)
		conn, err := net.DialTimeout("tcp", smtpAddr, 2*time.Second)
		if err != nil {
			logger.Warn("[HEALTH] SMTP health check failed: %v", err)
			response.Checks["smtp"] = "degraded: " + err.Error()
		} else {
			conn.Close()
			response.Checks["smtp"] = "healthy"
		}
	}

	// Check immich if used (non-fatal)
	immichURL := os.Getenv("IMMICH_URL")
	immichToken := os.Getenv("IMMICH_KEY")
	if immichURL != "" && immichToken != "" {
		client := immich.NewClient(immichURL, immichToken)
		if err := client.TestConnection(); err != nil {
			logger.Warn("[HEALTH] Immich health check failed: %v", err)
			response.Checks["immich"] = "degraded: " + err.Error()
		} else {
			response.Checks["immich"] = "healthy"
		}
	}

	statusCode := http.StatusOK
	if response.Status == "error" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
