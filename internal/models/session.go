/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package models

import (
	"fmt"
	"time"
)

// Session represents an active user session
type Session struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Token  string `json:"-"` // Don't expose in JSON

	// Device/Browser Info
	UserAgent  string `json:"user_agent"`
	Browser    string `json:"browser"`
	BrowserVer string `json:"browser_version"`
	OS         string `json:"os"`
	Device     string `json:"device"`
	IsMobile   bool   `json:"is_mobile"`

	// Network Info
	IPAddress string `json:"ip_address"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`

	// Session Metadata
	LoginTime    time.Time `json:"login_time"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`

	// Additional
	Referer  string `json:"referer,omitempty"`
	Language string `json:"language,omitempty"`

	// UI Helper (not in database)
	IsCurrent bool `json:"is_current"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// TimeAgo returns a human-readable string for last activity
func (s *Session) TimeAgo() string {
	duration := time.Since(s.LastActivity)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// ExpirationTimeLeft returns a human-readable string for time until expiration
func (s *Session) ExpirationTimeLeft() string {
	duration := time.Until(s.ExpiresAt)

	// If already expired
	if duration < 0 {
		return "expired"
	}

	switch {
	case duration < time.Minute:
		return "less than a minute"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	default:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// GetDeviceIcon returns an emoji icon for the device type
func (s *Session) GetDeviceIcon() string {
	if s.IsMobile {
		return "📱"
	}
	return "🖥️"
}

// GetHumanReadableDevice returns a friendly device description
func (s *Session) GetHumanReadableDevice() string {
	if s.Browser == "" || s.Browser == "Unknown" {
		return s.Device
	}
	return s.Browser + " on " + s.OS
}

// GetLocationString returns a formatted location string
func (s *Session) GetLocationString() string {
	if s.City != "" && s.Country != "" {
		return s.City + ", " + s.Country
	}
	if s.Country != "" {
		return s.Country
	}
	return "Unknown Location"
}
