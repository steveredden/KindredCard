package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
)

// Client represents an Immich API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Immich API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PingPong
type ServerPingPong struct {
	Result string `json:"res"`
}

// Person represents a person in Immich
type Person struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BirthDate     *string   `json:"birthDate,omitempty"`
	ThumbnailPath string    `json:"thumbnailPath"`
	IsHidden      bool      `json:"isHidden"`
	IsFavorite    bool      `json:"isFavorite"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ===== API METHODS =====

// TestConnection tests the connection to Immich
func (c *Client) TestConnection() error {
	logger.Debug("[IMMICH] Testing connection to %s", c.BaseURL)

	req, err := http.NewRequest("GET", c.BaseURL+"/api/server/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result ServerPingPong

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("[IMMICH] Failed to decode response: %v", err)
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Result == "pong" {
		logger.Info("[IMMICH] Connection successful")
		return nil
	} else {
		return fmt.Errorf("did not get expected 'pong'")
	}
}

// GetAllPeople retrieves all people from Immich
func (c *Client) GetAllPeople() ([]Person, error) {
	logger.Debug("[IMMICH] Fetching all people")

	req, err := http.NewRequest("GET", c.BaseURL+"/api/people", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		People []Person `json:"people"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Info("[IMMICH] Found %d people", len(result.People))
	return result.People, nil
}

// GetPerson retrieves a single person by ID
func (c *Client) GetPerson(personID string) (*Person, error) {
	logger.Debug("[IMMICH] Fetching person %s", personID)

	req, err := http.NewRequest("GET", c.BaseURL+"/api/people/"+personID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var person Person
	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &person, nil
}

// GetPersonThumbnail downloads a person's face thumbnail
func (c *Client) GetPersonThumbnail(personID string) ([]byte, error) {
	logger.Debug("[IMMICH] Downloading thumbnail for person %s", personID)

	url := fmt.Sprintf("%s/api/people/%s/thumbnail", c.BaseURL, personID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug("[IMMICH] Downloaded %d bytes", len(data))
	return data, nil
}

// UpdatePerson updates a person's information
func (c *Client) UpdatePerson(personID string, updates Person) error {
	logger.Debug("[IMMICH] Updating person %s", personID)

	url := fmt.Sprintf("%s/api/person/%s", c.BaseURL, personID)

	jsonData, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("[IMMICH] Updated person successfully")
	return nil
}
