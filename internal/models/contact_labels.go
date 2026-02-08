package models

// ContactLabelType represents the lookup table for labels (home, work, etc.)
type ContactLabelType struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Category   string `json:"category"` // 'phone', 'email', 'address', etc.
	IsSystem   bool   `json:"is_system"`
	UsageCount int    `json:"usage_count"`
}

type ContactLabelJSONPost struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}
