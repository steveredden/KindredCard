package models

// Email represents an email address
type Email struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	Email     string `json:"email"`
	Type      int    `json:"label_type_id"`
	TypeLabel string `json:"type_label"`
	IsPrimary bool   `json:"is_primary"`
}

type EmailJSONPatch struct {
	ID        int     `json:"id" example:"1"`
	ContactID *int    `json:"contact_id" example:"4"`
	Email     *string `json:"phone" example:"support@kindredcard.com"`
	Type      *int    `json:"label_type_id" example:"42"`
	IsPrimary *bool   `json:"is_primary" example:"false"`
}
