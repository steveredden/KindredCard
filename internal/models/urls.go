package models

// URL represents a website or online profile
type URL struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	URL       string `json:"url"`
	Type      int    `json:"label_type_id"`
	TypeLabel string `json:"type_label"`
	IsPrimary bool   `json:"is_primary"`
}

type URLJSONPatch struct {
	ID        int     `json:"id" example:"1"`
	ContactID *int    `json:"contact_id" example:"4"`
	URL       *string `json:"phone" example:"https://facebook.com/kindredcard"`
	Type      *int    `json:"label_type_id" example:"42"`
	IsPrimary *bool   `json:"is_primary" example:"false"`
}
