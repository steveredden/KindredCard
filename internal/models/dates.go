package models

import "time"

// Other Dated Events
type OtherDate struct {
	ID             int        `json:"id"`
	ContactID      int        `json:"contact_id"`
	EventName      string     `json:"event_name"`
	EventDate      *time.Time `json:"event_date,omitempty"`
	EventDateMonth *int       `json:"event_date_month,omitempty"` // 1-12, for partial dates
	EventDateDay   *int       `json:"event_date_day,omitempty"`   // 1-31, for partial dates
}
