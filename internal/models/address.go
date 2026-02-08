package models

// Address represents a physical address
type Address struct {
	ID             int    `json:"id"`
	ContactID      int    `json:"contact_id"`
	Street         string `json:"street"`
	ExtendedStreet string `json:"extended_street"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	Country        string `json:"country"`
	Type           int    `json:"label_type_id"`
	TypeLabel      string `json:"type_label"`
	IsPrimary      bool   `json:"is_primary"`
}

type AddressJSONPatch struct {
	ID             int     `json:"id" example:"1"`
	ContactID      *int    `json:"contact_id" example:"4"`
	Street         *string `json:"street"`
	ExtendedStreet *string `json:"extended_street"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postal_code"`
	Country        *string `json:"country"`
	Type           *int    `json:"label_type_id" example:"42"`
	IsPrimary      *bool   `json:"is_primary" example:"false"`
}

// Suggestion defines the proposed action for the UI
type AddressSuggestion struct {
	Type               string // "Address"
	TargetID           int    // The ID of the person we are ADDING the link to
	TargetName         string
	ProposedID         int // The ID of the person they are related to
	SourceName         string
	RelationshipTypeID int     // The ID for "Brother", "Father", etc.
	ProposedVal        Address `json:"proposed_val"`
	DisplayVal         string
	Reason             string // Your logic description
}
