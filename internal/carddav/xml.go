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
	"encoding/xml"
)

// XML Namespaces
const (
	NamespaceDAV     = "DAV:"
	NamespaceCardDAV = "urn:ietf:params:xml:ns:carddav"
)

// ========================================
// PROPFIND Request/Response Structures
// ========================================

type Propfind struct {
	XMLName xml.Name  `xml:"DAV: propfind"`
	Prop    *Prop     `xml:"prop,omitempty"`
	AllProp *struct{} `xml:"allprop,omitempty"`
}

type Prop struct {
	XMLName                xml.Name                `xml:"DAV: prop"`
	ResourceType           *ResourceType           `xml:"resourcetype,omitempty"`
	DisplayName            *DisplayName            `xml:"displayname,omitempty"`
	GetETag                *GetETag                `xml:"getetag,omitempty"`
	GetContentType         *GetContentType         `xml:"getcontenttype,omitempty"`
	CurrentUserPrincipal   *CurrentUserPrincipal   `xml:"current-user-principal,omitempty"`
	PrincipalURL           *PrincipalURL           `xml:"principal-URL,omitempty"`
	AddressBookHomeSet     *AddressBookHomeSet     `xml:"urn:ietf:params:xml:ns:carddav addressbook-home-set,omitempty"`
	SupportedReportSet     *SupportedReportSet     `xml:"supported-report-set,omitempty"`
	AddressData            *AddressData            `xml:"urn:ietf:params:xml:ns:carddav address-data,omitempty"`
	SyncToken              *SyncToken              `xml:"sync-token,omitempty"`
	AddressBookDescription *AddressBookDescription `xml:"urn:ietf:params:xml:ns:carddav addressbook-description,omitempty"`
}

type Multistatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []Response `xml:"response"`
	SyncToken string     `xml:"sync-token,omitempty"`
}

type Response struct {
	XMLName  xml.Name   `xml:"DAV: response"`
	Href     string     `xml:"href"`
	Propstat []Propstat `xml:"propstat,omitempty"`
	Status   string     `xml:"status,omitempty"` // For deleted items (404)
}

type Propstat struct {
	XMLName xml.Name `xml:"DAV: propstat"`
	Prop    PropData `xml:"prop"`
	Status  string   `xml:"status"`
}

type PropData struct {
	XMLName                xml.Name                `xml:"DAV: prop"`
	ResourceType           *ResourceType           `xml:"resourcetype,omitempty"`
	DisplayName            *DisplayName            `xml:"displayname,omitempty"`
	GetETag                *GetETag                `xml:"getetag,omitempty"`
	GetContentType         *GetContentType         `xml:"getcontenttype,omitempty"`
	CurrentUserPrincipal   *CurrentUserPrincipal   `xml:"current-user-principal,omitempty"`
	PrincipalURL           *PrincipalURL           `xml:"principal-URL,omitempty"`
	AddressBookHomeSet     *AddressBookHomeSet     `xml:"urn:ietf:params:xml:ns:carddav addressbook-home-set,omitempty"`
	SupportedReportSet     *SupportedReportSet     `xml:"supported-report-set,omitempty"`
	AddressData            *AddressData            `xml:"urn:ietf:params:xml:ns:carddav address-data,omitempty"`
	SyncToken              *SyncToken              `xml:"sync-token,omitempty"`
	AddressBookDescription *AddressBookDescription `xml:"urn:ietf:params:xml:ns:carddav addressbook-description,omitempty"`
}

// ========================================
// REPORT Request/Response Structures
// ========================================

type AddressBookMultiget struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav addressbook-multiget"`
	Prop    Prop     `xml:"DAV: prop"`
	Hrefs   []Href   `xml:"DAV: href"`
}

type AddressBookQuery struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav addressbook-query"`
	Prop    Prop     `xml:"DAV: prop"`
	Filter  *Filter  `xml:"urn:ietf:params:xml:ns:carddav filter,omitempty"`
}

type SyncCollection struct {
	XMLName   xml.Name  `xml:"DAV: sync-collection"`
	SyncToken SyncToken `xml:"sync-token"`
	SyncLevel string    `xml:"sync-level"`
	Prop      Prop      `xml:"prop"`
}

type Filter struct {
	XMLName     xml.Name     `xml:"urn:ietf:params:xml:ns:carddav filter"`
	PropFilters []PropFilter `xml:"prop-filter,omitempty"`
}

type PropFilter struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav prop-filter"`
	Name    string   `xml:"name,attr"`
}

// ========================================
// Property Elements
// ========================================

type ResourceType struct {
	XMLName     xml.Name     `xml:"DAV: resourcetype"`
	Collection  *Collection  `xml:"collection,omitempty"`
	Principal   *Principal   `xml:"principal,omitempty"`
	AddressBook *AddressBook `xml:"urn:ietf:params:xml:ns:carddav addressbook,omitempty"`
}

type Collection struct {
	XMLName xml.Name `xml:"DAV: collection"`
}

type Principal struct {
	XMLName xml.Name `xml:"DAV: principal"`
}

type AddressBook struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav addressbook"`
}

type DisplayName struct {
	XMLName xml.Name `xml:"DAV: displayname"`
	Value   string   `xml:",chardata"`
}

type GetETag struct {
	XMLName xml.Name `xml:"DAV: getetag"`
	Value   string   `xml:",chardata"`
}

type GetContentType struct {
	XMLName xml.Name `xml:"DAV: getcontenttype"`
	Value   string   `xml:",chardata"`
}

type CurrentUserPrincipal struct {
	XMLName xml.Name `xml:"DAV: current-user-principal"`
	Href    string   `xml:"href"`
}

type PrincipalURL struct {
	XMLName xml.Name `xml:"DAV: principal-URL"`
	Href    string   `xml:"href"`
}

type AddressBookHomeSet struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav addressbook-home-set"`
	Href    string   `xml:"DAV: href"`
}

type SupportedReportSet struct {
	XMLName          xml.Name          `xml:"DAV: supported-report-set"`
	SupportedReports []SupportedReport `xml:"supported-report"`
}

type SupportedReport struct {
	XMLName xml.Name `xml:"DAV: supported-report"`
	Report  Report   `xml:"report"`
}

type Report struct {
	XMLName             xml.Name             `xml:"DAV: report"`
	AddressBookMultiget *AddressBookMultiget `xml:"urn:ietf:params:xml:ns:carddav addressbook-multiget,omitempty"`
	AddressBookQuery    *AddressBookQuery    `xml:"urn:ietf:params:xml:ns:carddav addressbook-query,omitempty"`
}

type AddressData struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav address-data"`
	Value   string   `xml:",chardata"`
}

type SyncToken struct {
	XMLName xml.Name `xml:"DAV: sync-token"`
	Value   string   `xml:",chardata"`
}

type AddressBookDescription struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:carddav addressbook-description"`
	Value   string   `xml:",chardata"`
}

type Href struct {
	XMLName xml.Name `xml:"DAV: href"`
	Value   string   `xml:",chardata"`
}
