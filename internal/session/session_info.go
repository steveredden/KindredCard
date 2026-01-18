/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package session

import (
	"net"
	"net/http"
	"strings"
	"time"
)

// SessionInfo contains useful information about a user's session
type SessionInfo struct {
	// Device/Browser
	UserAgent  string
	Browser    string
	BrowserVer string
	OS         string
	Device     string
	IsMobile   bool

	// Network
	IPAddress string

	// Metadata
	LoginTime time.Time
	Referer   string
	Language  string
}

// GetSessionInfo extracts all useful session information from an HTTP request
func GetSessionInfo(r *http.Request) *SessionInfo {
	info := &SessionInfo{
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: GetClientIP(r),
		LoginTime: time.Now(),
		Referer:   r.Header.Get("Referer"),
		Language:  r.Header.Get("Accept-Language"),
	}

	// Parse user agent
	info.Browser, info.BrowserVer = ParseBrowser(info.UserAgent)
	info.OS = ParseOS(info.UserAgent)
	info.Device = ParseDevice(info.UserAgent)
	info.IsMobile = IsMobileDevice(info.UserAgent)

	return info
}

// GetClientIP extracts the real client IP address from the request
// Handles X-Forwarded-For and X-Real-IP headers for proxies
func GetClientIP(r *http.Request) string {
	// Try X-Forwarded-For first (comma-separated if multiple proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take first IP (original client)
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Try X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fallback to RemoteAddr (remove port)
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// ParseBrowser extracts browser name and version from User-Agent
func ParseBrowser(ua string) (name, version string) {
	switch {
	case strings.Contains(ua, "Edg/"):
		return "Edge", extractVersion(ua, "Edg/")
	case strings.Contains(ua, "Chrome/"):
		return "Chrome", extractVersion(ua, "Chrome/")
	case strings.Contains(ua, "Firefox/"):
		return "Firefox", extractVersion(ua, "Firefox/")
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome"):
		return "Safari", extractVersion(ua, "Version/")
	case strings.Contains(ua, "OPR/"):
		return "Opera", extractVersion(ua, "OPR/")
	default:
		return "Unknown", ""
	}
}

// ParseOS extracts operating system from User-Agent
func ParseOS(ua string) string {
	switch {
	case strings.Contains(ua, "Windows NT 10.0"):
		return "Windows 10/11"
	case strings.Contains(ua, "Windows NT"):
		return "Windows"
	case strings.Contains(ua, "Mac OS X"):
		return "macOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		return "iOS"
	default:
		return "Unknown"
	}
}

// ParseDevice extracts device type from User-Agent
func ParseDevice(ua string) string {
	switch {
	case strings.Contains(ua, "iPhone"):
		return "iPhone"
	case strings.Contains(ua, "iPad"):
		return "iPad"
	case strings.Contains(ua, "Android") && strings.Contains(ua, "Mobile"):
		return "Android Phone"
	case strings.Contains(ua, "Android"):
		return "Android Tablet"
	case strings.Contains(ua, "Mobile"):
		return "Mobile Device"
	default:
		return "Desktop"
	}
}

// IsMobileDevice checks if the request is from a mobile device
func IsMobileDevice(ua string) bool {
	return strings.Contains(ua, "Mobile") ||
		strings.Contains(ua, "Android") ||
		strings.Contains(ua, "iPhone") ||
		strings.Contains(ua, "iPad")
}

// GetHumanReadableDevice returns a user-friendly device description
func (s *SessionInfo) GetHumanReadableDevice() string {
	if s.Browser == "Unknown" {
		return s.Device
	}
	return s.Browser + " on " + s.OS
}

// GetDeviceIcon returns an appropriate icon/emoji for the device
func (s *SessionInfo) GetDeviceIcon() string {
	switch {
	case strings.Contains(s.Device, "iPhone"):
		return "📱"
	case strings.Contains(s.Device, "iPad"):
		return "📱"
	case strings.Contains(s.Device, "Android"):
		return "📱"
	case s.IsMobile:
		return "📱"
	default:
		return "🖥️"
	}
}

// extractVersion extracts version number after a prefix in User-Agent
func extractVersion(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx == -1 {
		return ""
	}

	start := idx + len(prefix)
	end := start

	// Extract version (numbers and dots)
	for end < len(ua) && (ua[end] >= '0' && ua[end] <= '9' || ua[end] == '.') {
		end++
	}

	return ua[start:end]
}
