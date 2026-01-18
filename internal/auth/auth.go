/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrMalformedPayload = errors.New("malformed token payload; missing user ID")
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against a hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateSignedToken generates a token that includes the UserID in the payload
// and is protected by an HMAC signature.
// Returns: "[UserID]-[RandomPart].[Signature]"
func GenerateSignedToken(appKey string, userID int) (string, error) {
	randomPart, err := GenerateSecureToken()
	if err != nil {
		return "", err
	}

	// 1. CONSTRUCT THE UNSIGNED PAYLOAD
	// The payload now includes the user ID for immediate identification.
	payload := fmt.Sprintf("%d-%s", userID, randomPart)

	// 2. SIGN THE ENTIRE PAYLOAD
	signature := signToken(payload, appKey)

	// 3. COMBINE PAYLOAD AND SIGNATURE
	return payload + "." + signature, nil
}

// VerifyAndExtractUserID takes the full signed token, verifies its signature,
// and extracts the user ID from the payload.
// Returns: (tokenPayload string, userID int, err error)
func VerifyAndExtractUserID(signedToken string, appKey string) (string, int, error) {
	// 1. Split the signed token into payload (token) and signature.

	// Find the last dot separator
	dotIndex := -1
	for i := len(signedToken) - 1; i >= 0; i-- {
		if signedToken[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		// Renaming the error type for clarity based on the new structure
		return "", 0, ErrInvalidToken
	}

	tokenPayload := signedToken[:dotIndex] // This is the payload: "[UserID]-[RandomPart]"
	signedSignature := signedToken[dotIndex+1:]

	// 2. Verify signature
	expectedSig := signToken(tokenPayload, appKey) // Use the full payload for signing
	if !CompareSecureStrings(signedSignature, expectedSig) {
		return "", 0, ErrInvalidSignature
	}

	// --- NEW LOGIC: Extract User ID from the Payload ---

	// 3. Extract the User ID from the token payload string.
	// The expected format is: "[UserID]-[RandomPart]"
	payloadParts := strings.SplitN(tokenPayload, "-", 2)

	// Check if we have both a User ID part and a random part
	if len(payloadParts) != 2 {
		// This handles cases where a token might be the old format (just random string)
		// or a new malformed one.
		return "", 0, ErrMalformedPayload
	}

	// Parse the User ID (which is the first part)
	userID, err := strconv.Atoi(payloadParts[0])
	if err != nil {
		// If the UserID part is not an integer
		return "", 0, ErrMalformedPayload
	}

	// 4. Success: Return the full payload and the extracted user ID
	return tokenPayload, userID, nil
}

// signToken creates an HMAC-SHA256 signature of the token
func signToken(token, appKey string) string {
	h := hmac.New(sha256.New, []byte(appKey))
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// Argon2Params contains the parameters for Argon2id hashing
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns secure default parameters for Argon2id
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPasswordArgon2 creates an Argon2id hash of the password (alternative to bcrypt)
func HashPasswordArgon2(password string, params *Argon2Params) (string, error) {
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Encode as base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encoded := "$argon2id$v=19$m=" +
		string(rune(params.Memory)) + ",t=" +
		string(rune(params.Iterations)) + ",p=" +
		string(rune(params.Parallelism)) + "$" +
		b64Salt + "$" + b64Hash

	return encoded, nil
}

// CompareSecureStrings performs constant-time comparison of two strings
func CompareSecureStrings(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// IsTokenExpired checks if a token expiration time has passed
func IsTokenExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// GetSessionExpiry returns a future timestamp for session expiration
func GetSessionExpiry() time.Time {
	return time.Now().Add(30 * 24 * time.Hour) // 30 days
}
