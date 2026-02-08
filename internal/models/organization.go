package models

// Organization represents a company/organization affiliation
type Organization struct {
	ID           int    `json:"id"`
	ContactID    int    `json:"contact_id"`
	Name         string `json:"name"`
	PhoneticName string `json:"phonetic_name"`
	Title        string `json:"title"`
	Department   string `json:"department"`
	IsPrimary    bool   `json:"is_primary"` //not yet used
}

type OrganizationJSONPatch struct {
	ID           int     `json:"id" example:"1"`
	ContactID    *int    `json:"contact_id" example:"4"`
	Name         *string `json:"name"`
	PhoneticName *string `json:"phonetic_name"`
	Title        *string `json:"title"`
	Department   *string `json:"department"`
}
