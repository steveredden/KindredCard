package models

import "time"

// Phone represents a phone number
type Phone struct {
	ID            int        `json:"id"`
	ContactID     int        `json:"contact_id"`
	Phone         string     `json:"phone"`
	Type          int        `json:"label_type_id"`
	TypeLabel     string     `json:"type_label"`
	IsPrimary     bool       `json:"is_primary"`
	LastFormatted *time.Time `json:"last_formatted_at"`
}

type PhoneJSONPatch struct {
	ID        int     `json:"id" example:"1"`
	ContactID *int    `json:"contact_id" example:"4"`
	Phone     *string `json:"phone" example:"(555)122-4121"`
	Type      *int    `json:"label_type_id" example:"42"`
	IsPrimary *bool   `json:"is_primary" example:"false"`
}
