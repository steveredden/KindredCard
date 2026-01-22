/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package carddav

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/steveredden/KindredCard/internal/converter"
	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/middleware"
	"github.com/steveredden/KindredCard/internal/utils"
)

type Server struct {
	db            *db.Database
	ReadOnly      bool
	baseURL       string
	userPrincipal string
	userID        int
}

func NewServer(database *db.Database, readOnly bool) *Server {
	return &Server{
		db:       database,
		ReadOnly: readOnly,
	}
}

// ServeHTTP handles CardDAV requests
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if s.ReadOnly {
		switch r.Method {
		case "PUT", "DELETE", "PROPPATCH", "MKCOL":
			logger.Debug("[CARDDAV] Blocked %s attempt in One-Way mode from %s", r.Method, r.RemoteAddr)
			w.Header().Set("Allow", "OPTIONS, GET, HEAD, PROPFIND, REPORT")
			http.Error(w, "Server is in Read-Only mode", http.StatusMethodNotAllowed)
			return
		}
	}

	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	s.userID = user.ID
	s.userPrincipal = user.Email
	s.baseURL = getExternalBaseURL(r)

	if logger.GetLevel() == logger.TRACE {
		if ua := r.Header.Get("User-Agent"); ua != "" {
			logger.Trace("[CARDDAV] User-Agent: %s", ua)
		}
		if body, err := io.ReadAll(r.Body); err == nil {
			logger.Trace("[CARDDAV] %s %s", r.Method, r.URL.Path)
			if len(body) > 0 {
				logger.Trace("[CARDDAV] Body:\n%s", string(body))
			}
			// Restore body for handlers
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	switch r.Method {
	case "OPTIONS":
		s.handleOptions(w, r)
	case "PROPFIND":
		s.handlePropfind(w, r)
	case "REPORT":
		s.handleReport(w, r)
	case "GET":
		s.handleGet(w, r)
	case "PUT":
		s.handlePut(w, r)
	case "DELETE":
		s.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("DAV", "1, 2, 3, addressbook")

	methods := []string{"OPTIONS", "GET", "HEAD", "PROPFIND", "REPORT"}

	// allow two way sync, eg, edit a contact in Apple Contacts -> it flows to server
	if !s.ReadOnly {
		methods = append(methods, "PUT", "DELETE", "POST")
	}

	allowHeader := strings.Join(methods, ", ")
	w.Header().Set("Allow", allowHeader)
	w.Header().Set("Public", allowHeader)
	w.WriteHeader(http.StatusOK)
}

// ========================================
// PROPFIND Handler - Routes based on REQUESTED PROPERTIES
//		CurrentUserPrincipal
//		AddressBookHomeSet
//		GetETag
// ========================================

func (s *Server) handlePropfind(w http.ResponseWriter, r *http.Request) {
	depth := r.Header.Get("Depth")
	if depth == "" {
		depth = "0"
	}

	// Parse XML to determine what client is asking for
	var propfindReq Propfind
	if err := xml.NewDecoder(r.Body).Decode(&propfindReq); err != nil {
		logger.Error("[CARDDAV] [PROPFIND] XML parse error: %v", err)
		http.Error(w, "Invalid XML", http.StatusBadRequest)
		return
	}

	// Route based on WHAT PROPERTIES ARE REQUESTED, not the path
	if propfindReq.Prop != nil {
		prop := propfindReq.Prop

		// Is client asking for principal properties?
		if prop.CurrentUserPrincipal != nil || prop.PrincipalURL != nil {
			logger.Debug("[CARDDAV] [PROPFIND] -> Principal discovery")
			s.respondPrincipal(w, r)
			return
		}

		// Is client asking for addressbook-home-set?
		if prop.AddressBookHomeSet != nil {
			logger.Debug("[CARDDAV] [PROPFIND] -> Addressbook home set")
			s.respondAddressbookHome(w, r)
			return
		}

		// Is client asking for collection properties (sync-token, resourcetype with addressbook)?
		if prop.SyncToken != nil || prop.ResourceType != nil {
			// Check path to determine if it's the collection or a contact
			if strings.HasSuffix(r.URL.Path, ".vcf") || strings.HasSuffix(r.URL.Path, ".vcf/") {
				logger.Debug("[CARDDAV] [PROPFIND] -> Individual contact")
				s.respondContact(w, r)
			} else {
				logger.Debug("[CARDDAV] [PROPFIND] -> Address book collection")
				s.respondCollection(w, r, depth)
			}
			return
		}

		// Is client asking for etag on collection (list contacts)?
		if prop.GetETag != nil {
			logger.Debug("[CARDDAV] [PROPFIND] -> ETags (contact list)")
			s.respondCollection(w, r, depth)
			return
		}
	}

	// Fallback: determine by path
	logger.Debug("[CARDDAV] [PROPFIND] -> Fallback to path routing")
	if strings.HasSuffix(r.URL.Path, ".vcf") || strings.HasSuffix(r.URL.Path, ".vcf/") {
		s.respondContact(w, r)
	} else if strings.Contains(r.URL.Path, "/contacts") {
		s.respondCollection(w, r, depth)
	} else {
		s.respondPrincipal(w, r)
	}
}

// ========================================
// REPORT Handler - Routes based on XML ROOT ELEMENT
// ========================================

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Peek at XML to determine report type by ROOT ELEMENT
	decoder := xml.NewDecoder(bytes.NewReader(bodyBytes))

	isAppleClient := isAppleCardDAVClient(r)

	// Read first token to get root element
	for {
		token, err := decoder.Token()
		if err != nil {
			http.Error(w, "Invalid XML", http.StatusBadRequest)
			return
		}

		if startElem, ok := token.(xml.StartElement); ok {
			// Route based on root element name
			switch startElem.Name.Local {
			case "sync-collection":
				logger.Debug("[CARDDAV] [REPORT] -> sync-collection")
				var syncReq SyncCollection
				if err := xml.Unmarshal(bodyBytes, &syncReq); err != nil {
					logger.Error("[CARDDAV] [REPORT] Parse error: %v", err)
					http.Error(w, "Invalid XML", http.StatusBadRequest)
					return
				}
				s.respondSyncCollection(w, syncReq)
				return

			case "addressbook-multiget":
				logger.Debug("[CARDDAV] [REPORT] -> addressbook-multiget")
				var multigetReq AddressBookMultiget
				if err := xml.Unmarshal(bodyBytes, &multigetReq); err != nil {
					logger.Error("[CARDDAV] [REPORT] Parse error: %v", err)
					http.Error(w, "Invalid XML", http.StatusBadRequest)
					return
				}
				s.respondAddressbookMultiget(w, multigetReq, isAppleClient)
				return

			case "addressbook-query":
				logger.Debug("[CARDDAV] [REPORT] -> addressbook-query")
				var queryReq AddressBookQuery
				if err := xml.Unmarshal(bodyBytes, &queryReq); err != nil {
					logger.Error("[CARDDAV] [REPORT] Parse error: %v", err)
					http.Error(w, "Invalid XML", http.StatusBadRequest)
					return
				}
				s.respondAddressbookQuery(w, queryReq)
				return

			default:
				logger.Error("[CARDDAV] [REPORT] Unknown report type: %s", startElem.Name.Local)
				http.Error(w, "Unsupported report type", http.StatusNotImplemented)
				return
			}
		}
	}
}

// ========================================
// GET, PUT, DELETE - Still use path (no XML to parse)
// ========================================

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	uid := extractUIDFromPath(r.URL.Path)

	contact, err := s.db.GetContactByUID(s.userID, uid, true)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	card := converter.ContactToVCard(contact, false)

	var buf bytes.Buffer
	encoder := vcard.NewEncoder(&buf)
	if err := encoder.Encode(card); err != nil {
		http.Error(w, "Error encoding vCard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, contact.ETag))
	w.Write(buf.Bytes())
}

// ========================================
// PUT Handler
// ========================================

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	decoder := vcard.NewDecoder(bytes.NewReader(body))
	card, err := decoder.Decode()
	if err != nil {
		http.Error(w, "Invalid vCard", http.StatusBadRequest)
		return
	}

	allContacts, _ := s.db.GetAllContactsAbbrv(s.userID, false)
	allRelTypes, _ := s.db.GetRelationshipTypes()

	contact, _ := converter.VCardToContact(card, allContacts, allRelTypes)
	uid := extractUIDFromPath(r.URL.Path)
	contact.UID = uid

	//Debug output vcard
	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[CARDDAV] Client->Server vCard PUT Body:")
		utils.Dump(card)
		logger.Trace("[CARDDAV] Contact Model post-translation:")
		utils.Dump(contact)
	}

	// Check if contact exists
	existing, err := s.db.GetContactByUID(s.userID, uid, true)
	if err == nil {
		// Update existing
		contact.ID = existing.ID
		if err := s.db.UpdateContact(s.userID, contact); err != nil {
			http.Error(w, "Error updating contact", http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", contact.ETag)
		w.WriteHeader(http.StatusNoContent)
	} else {
		// Create new
		if err := s.db.CreateContact(s.userID, contact); err != nil {
			http.Error(w, "Error creating contact", http.StatusInternalServerError)
			return
		}
		w.Header().Set("ETag", contact.ETag)
		w.WriteHeader(http.StatusCreated)
	}
}

// ========================================
// DELETE Handler
// ========================================

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, ".vcf"), "/")
	uid := parts[len(parts)-1]

	contactID, ok := s.db.GetContactIDByUID(s.userID, uid)
	if !ok {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	if err := s.db.DeleteContact(s.userID, contactID); err != nil {
		http.Error(w, "Error deleting contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========================================
// Response Generators
// ========================================

func (s *Server) respondPrincipal(w http.ResponseWriter, r *http.Request) {
	principalPath := fmt.Sprintf("/carddav/%s/", s.userPrincipal)
	principalURL := s.baseURL + principalPath

	response := Multistatus{
		Responses: []Response{
			{
				Href: principalURL,
				Propstat: []Propstat{
					{
						Prop: PropData{
							ResourceType: &ResourceType{
								Collection: &Collection{},
								Principal:  &Principal{},
							},
							CurrentUserPrincipal: &CurrentUserPrincipal{
								Href: principalURL,
							},
							PrincipalURL: &PrincipalURL{
								Href: principalURL,
							},
							DisplayName: &DisplayName{
								Value: s.userPrincipal,
							},
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			},
		},
	}

	s.writeXMLResponse(w, response)
}

func (s *Server) respondAddressbookHome(w http.ResponseWriter, r *http.Request) {
	principalPath := fmt.Sprintf("/carddav/%s/", s.userPrincipal)
	contactsPath := principalPath + "contacts/"

	response := Multistatus{
		Responses: []Response{
			{
				Href: s.baseURL + principalPath,
				Propstat: []Propstat{
					{
						Prop: PropData{
							AddressBookHomeSet: &AddressBookHomeSet{
								Href: s.baseURL + contactsPath,
							},
							SupportedReportSet: &SupportedReportSet{
								SupportedReports: []SupportedReport{
									{Report: Report{AddressBookMultiget: &AddressBookMultiget{}}},
									{Report: Report{AddressBookQuery: &AddressBookQuery{}}},
								},
							},
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			},
		},
	}

	s.writeXMLResponse(w, response)
}

func (s *Server) respondCollection(w http.ResponseWriter, r *http.Request, depth string) {
	collectionPath := fmt.Sprintf("/carddav/%s/contacts/", s.userPrincipal)
	currentToken, _ := s.db.GetAddressBookSyncToken(s.userID)
	tokenStr := strconv.Itoa(currentToken)

	responses := []Response{
		{
			Href: s.baseURL + collectionPath,
			Propstat: []Propstat{
				{
					Prop: PropData{
						ResourceType: &ResourceType{
							Collection:  &Collection{},
							AddressBook: &AddressBook{},
						},
						DisplayName: &DisplayName{
							Value: "Contacts",
						},
						SyncToken: &SyncToken{
							//Value: fmt.Sprintf("%s%stoken/%d", s.baseURL, collectionPath, currentToken),
							Value: tokenStr,
						},
						GetCTag: &GetCTag{
							Value: tokenStr,
						},
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		},
	}

	// If depth is 1, include all contacts
	if depth == "1" {
		contacts, _ := s.db.GetAllContactsAbbrv(s.userID, true)
		for _, contact := range contacts {
			responses = append(responses, Response{
				Href: s.baseURL + collectionPath + contact.UID + ".vcf",
				Propstat: []Propstat{
					{
						Prop: PropData{
							ResourceType:   &ResourceType{},
							GetETag:        &GetETag{Value: fmt.Sprintf(`"%s"`, contact.ETag)},
							GetContentType: &GetContentType{Value: "text/vcard; charset=utf-8"},
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			})
		}
	}

	s.writeXMLResponse(w, Multistatus{Responses: responses})
}

func (s *Server) respondContact(w http.ResponseWriter, r *http.Request) {
	uid := extractUIDFromPath(r.URL.Path)

	contact, err := s.db.GetContactByUID(s.userID, uid, true)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	collectionPath := fmt.Sprintf("/carddav/%s/contacts/", s.userPrincipal)

	response := Multistatus{
		Responses: []Response{
			{
				Href: s.baseURL + collectionPath + contact.UID + ".vcf",
				Propstat: []Propstat{
					{
						Prop: PropData{
							ResourceType:   &ResourceType{},
							GetETag:        &GetETag{Value: fmt.Sprintf(`"%s"`, contact.ETag)},
							GetContentType: &GetContentType{Value: "text/vcard; charset=utf-8"},
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			},
		},
	}

	s.writeXMLResponse(w, response)
}

func (s *Server) respondSyncCollection(w http.ResponseWriter, req SyncCollection) {
	currentToken, _ := s.db.GetAddressBookSyncToken(s.userID)

	// Extract client token from parsed XML (not from URL string!)
	clientTokenStr := extractTokenFromURL(req.SyncToken.Value)
	clientToken, _ := strconv.ParseInt(clientTokenStr, 10, 64)

	contacts, _ := s.db.ListContactsChangedSince(s.userID, clientToken, true)

	collectionPath := fmt.Sprintf("/carddav/%s/contacts/", s.userPrincipal)
	responses := []Response{}

	for _, contact := range contacts {
		hrefURL := s.baseURL + collectionPath + contact.UID + ".vcf"

		if contact.DeletedAt != nil {
			// Deleted contact
			responses = append(responses, Response{
				Href:   hrefURL,
				Status: "HTTP/1.1 404 Not Found",
			})
		} else {
			// Active or modified contact
			responses = append(responses, Response{
				Href: hrefURL,
				Propstat: []Propstat{
					{
						Prop: PropData{
							GetETag: &GetETag{Value: fmt.Sprintf(`"%s"`, contact.ETag)},
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			})
		}
	}

	// nextTokenURL := fmt.Sprintf("%s%stoken/%d", s.baseURL, collectionPath, currentToken)
	tokenStr := strconv.Itoa(currentToken)

	s.writeXMLResponse(w, Multistatus{
		Responses: responses,
		SyncToken: tokenStr,
	})
}

func (s *Server) respondAddressbookMultiget(w http.ResponseWriter, req AddressBookMultiget, isAppleClient bool) {
	collectionPath := fmt.Sprintf("/carddav/%s/contacts/", s.userPrincipal)
	wantsAddressData := req.Prop.AddressData != nil

	responses := []Response{}

	// Extract UIDs from parsed XML hrefs (not by string parsing!)
	for _, href := range req.Hrefs {
		uid := extractUIDFromHref(href.Value)

		contact, err := s.db.GetContactByUID(s.userID, uid, true)
		if err != nil {
			continue
		}

		propData := PropData{
			GetETag: &GetETag{Value: fmt.Sprintf(`"%s"`, contact.ETag)},
		}

		// Include vCard data if requested (determined by XML property presence!)
		if wantsAddressData {
			card := converter.ContactToVCard(contact, isAppleClient)
			var buf bytes.Buffer
			encoder := vcard.NewEncoder(&buf)
			encoder.Encode(card)

			//Trace output vcard
			if logger.GetLevel() == logger.TRACE {
				logger.Trace("[CARDDAV] Server->Client vCard REPORT Body:")
				utils.Dump(card)
				logger.Trace("[CARDDAV] Contact Model pre-translation:")
				utils.Dump(contact)
			}

			propData.AddressData = &AddressData{
				Value: buf.String(),
			}
		}

		responses = append(responses, Response{
			Href: s.baseURL + collectionPath + contact.UID + ".vcf",
			Propstat: []Propstat{
				{
					Prop:   propData,
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}

	s.writeXMLResponse(w, Multistatus{Responses: responses})
}

func (s *Server) respondAddressbookQuery(w http.ResponseWriter, req AddressBookQuery) {
	// For now, return all contacts (filtering can be added based on req.Filter)
	contacts, _ := s.db.GetAllContacts(s.userID, false)

	collectionPath := fmt.Sprintf("/carddav/%s/contacts/", s.userPrincipal)
	wantsAddressData := req.Prop.AddressData != nil

	responses := []Response{}

	for _, contact := range contacts {
		propData := PropData{
			GetETag: &GetETag{Value: fmt.Sprintf(`"%s"`, contact.ETag)},
		}

		if wantsAddressData {
			card := converter.ContactToVCard(contact, false)
			var buf bytes.Buffer
			encoder := vcard.NewEncoder(&buf)
			encoder.Encode(card)

			propData.AddressData = &AddressData{
				Value: buf.String(),
			}
		}

		responses = append(responses, Response{
			Href: s.baseURL + collectionPath + contact.UID + ".vcf",
			Propstat: []Propstat{
				{
					Prop:   propData,
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}

	s.writeXMLResponse(w, Multistatus{Responses: responses})
}

// ========================================
// Helper Functions
// ========================================

func (s *Server) writeXMLResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	w.Write([]byte(xml.Header))

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(v); err != nil {
		logger.Error("[CARDDAV] XML Encoding error: %v", err)
	}
}

func extractUIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasSuffix(parts[i], ".vcf") {
			return strings.TrimSuffix(parts[i], ".vcf")
		}
	}
	return ""
}

func extractTokenFromURL(tokenURL string) string {
	parts := strings.Split(tokenURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractUIDFromHref(href string) string {
	// URL decode first
	href = strings.ReplaceAll(href, "%40", "@")

	parts := strings.Split(href, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasSuffix(parts[i], ".vcf") {
			return strings.TrimSuffix(parts[i], ".vcf")
		}
	}
	return ""
}

func getExternalBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = fwd
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

func isAppleCardDAVClient(r *http.Request) bool {
	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return false
	}

	// Check for strong indicators of an Apple CardDAV client
	upperUA := strings.ToUpper(ua)
	return strings.Contains(upperUA, "IOS") || strings.Contains(upperUA, "MACOS")
}
