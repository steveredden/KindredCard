package models

// IMPP represents a Instant Messaging profile/protocol
type IMPP struct {
	ID        int    `json:"id"`
	ContactID int    `json:"contact_id"`
	IMPP      string `json:"impp"`
	Type      int    `json:"label_type_id"`
	TypeLabel string `json:"type_label"`
}

type IMPPJSONPatch struct {
	ID        int     `json:"id" example:"1"`
	ContactID *int    `json:"contact_id" example:"4"`
	IMPP      *string `json:"impp" example:"matrix:u/john:example.org"`
	Type      *int    `json:"label_type_id" example:"42"`
}
