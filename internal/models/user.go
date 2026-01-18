/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package models

import "time"

// User represents an authenticated user
type User struct {
	ID              int       `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	IsSetupComplete bool      `json:"is_setup_complete"`
	Theme           string    `json:"theme"`
	Timezone        string    `json:"timezone"`
	SyncToken       int       `json:"addressbook_sync_token"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
