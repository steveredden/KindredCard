/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package converter

import (
	"fmt"
	"strings"

	"github.com/emersion/go-vcard"
)

// photoParts holds the extracted MIME type and Base64 content.
type photoParts struct {
	MimeType   string
	Base64Data string
}

// ParseVCardPhotoProperty extracts the MIME type and Base64 data from a vCard PHOTO field.
// It uses hardcoded parameter keys to avoid reliance on unexported vcard library constants.
func parseVCardPhotoProperty(field *vcard.Field) (photoParts, error) {
	propertyValue := field.Value
	upperValue := strings.ToUpper(propertyValue)

	// --- 1. Get Encoding and Type from Field Parameters using hardcoded keys ---
	// Hardcoded constants for parameters not exported by the vcard library version
	const (
		ParamEncoding  = "ENCODING"
		ParamType      = "TYPE"
		ParamValue     = "VALUE"
		ValueURI       = "URI"
		EncodingBase64 = "B"
	)

	encoding := strings.ToUpper(field.Params.Get(ParamEncoding))
	mimeTypeParam := strings.ToUpper(field.Params.Get(ParamType))
	valueType := strings.ToUpper(field.Params.Get(ParamValue))

	// --- 2. Handle VCard 4.0/Data URL Format (VALUE=URI) ---
	if valueType == ValueURI {
		// Check if it's a data URL, not a regular URL
		if !strings.Contains(upperValue, "DATA:") {
			return photoParts{}, fmt.Errorf("vcard photo: VALUE=URI found, but not a data URL (missing 'data:' prefix)")
		}

		cleanValue := strings.ReplaceAll(propertyValue, "\\,", ",")

		// Expected format: data:<mime-type>;base64,<data>
		dataStart := strings.Index(cleanValue, "data:")
		commaIndex := strings.Index(cleanValue, ",")

		if dataStart == -1 || commaIndex == -1 || commaIndex < dataStart {
			return photoParts{}, fmt.Errorf("vcard photo: invalid Data URL format (missing separator)")
		}

		// Extract MIME Type from the data URL part: data:<MIME>;base64
		mimeTypeSection := cleanValue[dataStart+5 : commaIndex]
		mimeType := strings.Split(mimeTypeSection, ";")[0] // Get 'image/jpeg' from 'image/jpeg;base64'

		base64Data := cleanValue[commaIndex+1:]

		return photoParts{
			MimeType:   strings.ToLower(mimeType),
			Base64Data: strings.TrimSpace(base64Data),
		}, nil
	}

	// --- 3. Handle VCard 3.0/B-Encoding Format (PHOTO;ENCODING=b;...) ---
	if encoding == EncodingBase64 || encoding == "BASE64" {

		// Find the colon that separates the metadata from the data
		// Use LastIndex to skip proprietary parameters (VND-63-MEMOJI-DETAILS)
		colonIndex := strings.LastIndex(propertyValue, ":")

		var base64Data string
		if colonIndex != -1 {
			// Data starts after the last colon
			base64Data = propertyValue[colonIndex+1:]
		} else {
			// Fallback: Use the whole value if no colon is found, common with line folding
			base64Data = propertyValue
		}

		base64Data = strings.TrimSpace(base64Data) // Crucial for handling line folding whitespace

		// --- Determine MIME Type ---
		var mimeType string
		if mimeTypeParam != "" {
			// Use the parameter value first (e.g., TYPE=PNG)
			mimeType = "image/" + strings.ToLower(mimeTypeParam)
			if mimeType == "image/jpg" { // Normalize JPG to JPEG
				mimeType = "image/jpeg"
			}
		}

		// --- Fallback Logic: Check Base64 Signature (Handles HEIC) ---
		if mimeType == "" || mimeType == "image/HEIC" {
			// The first 16 characters of HEIC Base64 data are a known signature
			const heicSignaturePrefix = "AAAAJGZ0eXBoZWlj"
			if strings.HasPrefix(base64Data, heicSignaturePrefix) {
				mimeType = "image/heic"
			}
		}

		if mimeType == "" {
			return photoParts{}, fmt.Errorf("vcard photo: B-encoding format missing required TYPE parameter or recognizable signature")
		}

		return photoParts{
			MimeType:   mimeType,
			Base64Data: base64Data,
		}, nil
	}

	return photoParts{}, fmt.Errorf("vcard photo: unrecognized photo property format or missing ENCODING=B")
}

// extractCustomLabel looks for a grouped X-ABLABEL and cleans it
func extractCustomLabel(card vcard.Card, group string) string {
	if group == "" {
		return ""
	}
	for _, label := range card[XLabelField] {
		if label.Group == group {
			// Strip Apple markers and return clean text
			clean := strings.TrimSuffix(strings.TrimPrefix(label.Value, "_$!<"), ">!$_")
			return strings.ToLower(clean)
		}
	}
	return ""
}

// addCustomLabel adds an X-ABLABEL field to the card linked to a group
func addCustomLabel(card vcard.Card, group string, label string) {
	if label == "" {
		return
	}

	labelField := &vcard.Field{
		Group: group,
		Value: label,
	}
	card.Add(XLabelField, labelField)
}
