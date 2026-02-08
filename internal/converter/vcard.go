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
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/emersion/go-vcard"
	"github.com/google/uuid"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

const (
	AppleOmitYearKey    = "X-APPLE-OMIT-YEAR"
	AppleOmitYearValue  = "1604"
	AppleOmitYearValueI = 1604
	AppleOmitYearPrefix = "1604-"

	XLabelField              = "X-ABLABEL"
	XDateField               = "X-ABDATE"
	XRelatedNamesField       = "X-ABRELATEDNAMES"
	XSocialProfileField      = "X-SOCIALPROFILE"
	XMaidenNameField         = "X-MAIDENNAME"
	XPhoneticFirstField      = "X-PHONETIC-FIRST-NAME"
	XPronunciationFirstField = "X-PRONUNCIATION-FIRST-NAME"
	XPhoneticLastField       = "X-PHONETIC-LAST-NAME"
	XPronunciationLastField  = "X-PRONUNCIATION-LAST-NAME"
	XPhoneticMiddleField     = "X-PHONETIC-MIDDLE-NAME"
	XPhoneticOrgField        = "X-PHONETIC-ORG"
)

// ContactToVCard converts a Contact model to a vCard
func ContactToVCard(contact *models.Contact, labelMap map[int]models.ContactLabelType, isAppleClient bool) vcard.Card {
	var extraItemIndex int = 1 //function-global extra item index

	card := make(vcard.Card)

	// Version
	card.SetValue(vcard.FieldVersion, "4.0")

	// UID
	card.SetValue(vcard.FieldUID, contact.UID)

	// Structured name (N field)
	name := &vcard.Name{
		FamilyName:      contact.FamilyName,
		GivenName:       contact.GivenName,
		AdditionalName:  contact.MiddleName,
		HonorificPrefix: contact.Prefix,
		HonorificSuffix: contact.Suffix,
	}
	card.SetName(name)

	// Name fields
	if contact.FullName != "" {
		card.SetValue(vcard.FieldFormattedName, contact.FullName)
	} else {
		card.SetValue(vcard.FieldFormattedName, contact.GenerateFullName())
	}

	// Nickname
	if contact.Nickname != "" {
		card.SetValue(vcard.FieldNickname, contact.Nickname)
	}

	// Maiden Name
	if contact.MaidenName != "" {
		card.Add(XMaidenNameField, &vcard.Field{Value: contact.MaidenName})
	}

	// Phonetics & Pronunciation
	if contact.PhoneticFirstName != "" {
		card.Add(XPhoneticFirstField, &vcard.Field{Value: contact.PhoneticFirstName})
	}
	if contact.PronunciationFirstName != "" {
		card.Add(XPronunciationFirstField, &vcard.Field{Value: contact.PronunciationFirstName})
	}
	if contact.PhoneticMiddleName != "" {
		card.Add(XPhoneticMiddleField, &vcard.Field{Value: contact.PhoneticMiddleName})
	}
	if contact.PhoneticLastName != "" {
		card.Add(XPhoneticLastField, &vcard.Field{Value: contact.PhoneticLastName})
	}
	if contact.PronunciationLastName != "" {
		card.Add(XPronunciationLastField, &vcard.Field{Value: contact.PronunciationLastName})
	}

	// Gender
	if contact.Gender != "" {
		card.SetValue(vcard.FieldGender, contact.Gender)
	}

	// Birthday - try full date first, then partial
	// https://datatracker.ietf.org/doc/html/rfc6350#section-6.2.5
	if contact.Birthday != nil {
		card.SetValue(vcard.FieldBirthday, contact.Birthday.Format("20060102"))
	} else if contact.BirthdayMonth != nil && contact.BirthdayDay != nil {
		field := &vcard.Field{}

		if isAppleClient {
			if field.Params == nil {
				field.Params = make(vcard.Params)
			}
			field.Params.Set(AppleOmitYearKey, AppleOmitYearValue)
			field.Value = fmt.Sprintf("%s-%02d-%02d", AppleOmitYearValue, *contact.BirthdayMonth, *contact.BirthdayDay)
		} else {
			// Partial birthday: --MMDD format
			field.Value = fmt.Sprintf("--%02d%02d", *contact.BirthdayMonth, *contact.BirthdayDay)
		}

		card.Add(vcard.FieldBirthday, field)
	}

	// Anniversary - try full date first, then partial
	// https://datatracker.ietf.org/doc/html/rfc6350#section-6.2.6
	if contact.Anniversary != nil {

		if isAppleClient {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			field := &vcard.Field{Value: contact.Anniversary.Format("2006-01-02")}
			field.Group = itemKey
			card.Add(XDateField, field)

			labelField := &vcard.Field{Value: "_$!<Anniversary>!$_"}
			labelField.Group = itemKey
			card.Add(XLabelField, labelField)
		} else {
			card.SetValue(vcard.FieldAnniversary, contact.Anniversary.Format("20060102"))
		}
	} else if contact.AnniversaryMonth != nil && contact.AnniversaryDay != nil {

		if isAppleClient {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			field := &vcard.Field{}
			field.Value = fmt.Sprintf("%s-%02d-%02d", AppleOmitYearValue, *contact.AnniversaryMonth, *contact.AnniversaryDay)
			if field.Params == nil {
				field.Params = make(vcard.Params)
			}
			field.Params.Set(AppleOmitYearKey, AppleOmitYearValue)
			field.Group = itemKey
			card.Add(XDateField, field)

			labelField := &vcard.Field{Value: "_$!<Anniversary>!$_"}
			labelField.Group = itemKey
			card.Add(XLabelField, labelField)
		} else {
			// Partial anniversary: --MMDD format
			card.SetValue(vcard.FieldAnniversary, fmt.Sprintf("--%02d%02d", *contact.AnniversaryMonth, *contact.AnniversaryDay))
		}
	}

	// Other Dates
	for _, otherDate := range contact.OtherDates {
		dateField := &vcard.Field{}
		labelField := &vcard.Field{}

		if otherDate.EventDate != nil {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			dateField.Group = itemKey
			labelField.Group = itemKey

			dateField.Value = otherDate.EventDate.Format("20060102")
			labelField.Value = otherDate.EventName

			card.Add(XLabelField, labelField)
			card.Add(XDateField, dateField)

		} else if otherDate.EventDateMonth != nil && otherDate.EventDateDay != nil {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			dateField.Group = itemKey
			labelField.Group = itemKey

			if isAppleClient {
				dateField.Value = fmt.Sprintf("%s-%02d-%02d", AppleOmitYearValue, *otherDate.EventDateMonth, *otherDate.EventDateDay)
				if dateField.Params == nil {
					dateField.Params = make(vcard.Params)
				}
				dateField.Params.Set(AppleOmitYearKey, AppleOmitYearValue)

			} else {
				dateField.Value = fmt.Sprintf("--%02d%02d", *otherDate.EventDateMonth, *otherDate.EventDateDay)
			}
			labelField.Value = otherDate.EventName

			card.Add(XLabelField, labelField)
			card.Add(XDateField, dateField)

		}

	}

	// Emails
	for _, email := range contact.Emails {

		field := &vcard.Field{
			Value:  email.Email,
			Params: make(vcard.Params),
		}

		if label, ok := labelMap[email.Type]; ok {
			if label.IsSystem {
				field.Params.Add(vcard.ParamType, label.Name)
			} else {
				itemKey := "item" + strconv.Itoa(extraItemIndex)
				extraItemIndex++

				field.Group = itemKey
				addCustomLabel(card, itemKey, label.Name) //add an X-ABLABEL custom label name
			}
		}

		if email.IsPrimary {
			field.Params.Set(vcard.ParamPreferred, "1")
			if isAppleClient {
				field.Params.Add(vcard.ParamType, "pref")
			}
		}
		card.Add(vcard.FieldEmail, field)
	}

	// Phones
	for _, phone := range contact.Phones {

		field := &vcard.Field{
			Value:  phone.Phone,
			Params: make(vcard.Params),
		}

		if label, ok := labelMap[phone.Type]; ok {
			if label.IsSystem {
				field.Params.Add(vcard.ParamType, label.Name)
			} else {
				itemKey := "item" + strconv.Itoa(extraItemIndex)
				extraItemIndex++

				field.Group = itemKey
				addCustomLabel(card, itemKey, label.Name) //add an X-ABLABEL custom label name
			}
		}

		if phone.IsPrimary {
			field.Params.Set(vcard.ParamPreferred, "1")
			if isAppleClient {
				field.Params.Add(vcard.ParamType, "pref")
			}
		}
		card.Add(vcard.FieldTelephone, field)
	}

	// Addresses
	for _, addr := range contact.Addresses {
		address := &vcard.Address{
			StreetAddress:   addr.Street,
			ExtendedAddress: addr.ExtendedStreet,
			Locality:        addr.City,
			Region:          addr.State,
			PostalCode:      addr.PostalCode,
			Country:         addr.Country,
			Field: &vcard.Field{
				Params: make(vcard.Params),
			},
		}

		if label, ok := labelMap[addr.Type]; ok {
			if label.IsSystem {
				address.Field.Params.Add(vcard.ParamType, label.Name)
			} else {
				itemKey := "item" + strconv.Itoa(extraItemIndex)
				extraItemIndex++

				address.Group = itemKey
				addCustomLabel(card, itemKey, label.Name) //add an X-ABLABEL custom label name
			}
		}

		if addr.IsPrimary {
			address.Field.Params.Set(vcard.ParamPreferred, "1")
			if isAppleClient {
				address.Field.Params.Add(vcard.ParamType, "pref")
			}
		}
		if addr.IsPrimary {
			address.Field.Params.Set(vcard.ParamPreferred, "1")
		}
		card.AddAddress(address)
	}

	// Organizations -- only do the first for now
	if len(contact.Organizations) > 0 {
		org := contact.Organizations[0]

		if org.Name != "" || org.Department != "" {
			orgValue := org.Name
			if org.Department != "" {
				orgValue += ";" + org.Department
			}
			card.SetValue(vcard.FieldOrganization, orgValue)
		}

		if org.Title != "" {
			card.SetValue(vcard.FieldTitle, org.Title)
		}

		if org.PhoneticName != "" {
			card.Add(XPhoneticOrgField, &vcard.Field{Value: org.PhoneticName})
		}
	}

	// URLs
	for _, url := range contact.URLs {
		field := &vcard.Field{
			Value:  url.URL,
			Params: make(vcard.Params),
		}

		if label, ok := labelMap[url.Type]; ok {
			if label.IsSystem {
				field.Params.Add(vcard.ParamType, label.Name)
			} else {
				itemKey := "item" + strconv.Itoa(extraItemIndex)
				extraItemIndex++

				field.Group = itemKey
				addCustomLabel(card, itemKey, label.Name) //add an X-ABLABEL custom label name
			}
		}

		if url.IsPrimary {
			field.Params.Set(vcard.ParamPreferred, "1")
			if isAppleClient {
				field.Params.Add(vcard.ParamType, "pref")
			}
		}
		card.Add(vcard.FieldURL, field)
	}

	// Notes
	if contact.Notes != "" {
		card.SetValue(vcard.FieldNote, contact.Notes)
	}

	// Avatar Photo
	if contact.AvatarBase64 != "" {

		photoField := &vcard.Field{Value: contact.AvatarBase64}
		photoParams := make(vcard.Params)

		// Determine the MIME subtype for the TYPE parameter
		mimeSubtype := strings.ToUpper(strings.TrimPrefix(contact.AvatarMimeType, "image/"))
		photoParams.Add(vcard.ParamType, mimeSubtype)

		if isAppleClient {
			// vCard 3.0 (B-Encoding)
			photoParams.Add("ENCODING", "b")
		} else {
			// vCard 4.0 (Data URI)
			photoParams.Add(vcard.ParamValue, "URI")
			dataURI := fmt.Sprintf("data:%s;base64,%s", contact.AvatarMimeType, contact.AvatarBase64)
			photoField.Value = dataURI
		}

		photoField.Params = photoParams
		card.Add(vcard.FieldPhoto, photoField)
	}

	// Relationships
	for _, rel := range contact.Relationships {
		if rel.RelatedContact != nil {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			namefield := &vcard.Field{Value: rel.RelatedContact.FullName}
			labelField := &vcard.Field{Value: rel.RelationshipType.Name}

			namefield.Group = itemKey
			labelField.Group = itemKey

			card.Add(XLabelField, labelField)
			card.Add(XRelatedNamesField, namefield)
		}
	}

	// Other Relationships
	for _, rel := range contact.OtherRelationships {
		if rel.RelatedContactName != "" && rel.RelationshipName != "" {
			itemKey := "item" + strconv.Itoa(extraItemIndex)
			extraItemIndex++

			namefield := &vcard.Field{Value: rel.RelatedContactName}
			labelField := &vcard.Field{Value: rel.RelationshipName}

			namefield.Group = itemKey
			labelField.Group = itemKey

			card.Add(XLabelField, labelField)
			card.Add(XRelatedNamesField, namefield)
		}
	}

	// Revision
	card.SetValue(vcard.FieldRevision, contact.UpdatedAt.Format(time.RFC3339))

	return card
}

// VCardToContact converts a vCard to a Contact model
func VCardToContact(card vcard.Card, allContacts []*models.Contact, allRelationshipTypes []models.RelationshipType, revMap map[string]int) (*models.Contact, error) {
	uid := ""
	if field := card.Get(vcard.FieldUID); field != nil && field.Value != "" {
		uid = field.Value
	} else {
		uid = uuid.New().String() // generate a new UUID
	}

	contact := &models.Contact{
		UID: uid,
	}

	// Name fields
	if fn := card.Get(vcard.FieldFormattedName); fn != nil {
		contact.FullName = fn.Value
	}

	if n := card.Name(); n != nil {
		contact.GivenName = n.GivenName
		contact.FamilyName = n.FamilyName
		contact.MiddleName = n.AdditionalName
		contact.Prefix = n.HonorificPrefix
		contact.Suffix = n.HonorificSuffix
	}

	// If FullName is empty, generate it
	if contact.FullName == "" {
		contact.FullName = contact.GenerateFullName()
	}

	// Nickname
	if nick := card.Get(vcard.FieldNickname); nick != nil {
		contact.Nickname = nick.Value
	}

	// Maiden Name
	if maiden := card.Get(XMaidenNameField); maiden != nil {
		contact.MaidenName = maiden.Value
	}

	// Phonetics & Pronunciation
	if phoneticFirst := card.Get(XPhoneticFirstField); phoneticFirst != nil {
		contact.PhoneticFirstName = phoneticFirst.Value
	}
	if pronuncFirst := card.Get(XPhoneticFirstField); pronuncFirst != nil {
		contact.PhoneticFirstName = pronuncFirst.Value
	}
	if phoneticMiddle := card.Get(XPhoneticFirstField); phoneticMiddle != nil {
		contact.PhoneticFirstName = phoneticMiddle.Value
	}
	if phoneticLast := card.Get(XPhoneticFirstField); phoneticLast != nil {
		contact.PhoneticFirstName = phoneticLast.Value
	}
	if pronuncLast := card.Get(XPhoneticFirstField); pronuncLast != nil {
		contact.PhoneticFirstName = pronuncLast.Value
	}

	// Gender
	if gender := card.Get(vcard.FieldGender); gender != nil {
		contact.Gender = gender.Value
	}

	// Birthday
	if bday := card.Get(vcard.FieldBirthday); bday != nil {
		birthday := bday.Value
		if strings.HasPrefix(birthday, "--") {
			// Partial date format: --MMDD
			if len(birthday) == 6 {
				month := utils.ParseIntPtr(birthday[2:4])
				day := utils.ParseIntPtr(birthday[4:6])
				contact.BirthdayMonth = month
				contact.BirthdayDay = day
			}
		} else {
			// Full date format: YYYYMMDD or YYYY-MM-DD
			t, _ := time.Parse("20060102", birthday)
			if t.IsZero() {
				t, _ = time.Parse("2006-01-02", birthday)
			}

			//Apple specific logic
			if t.Year() == AppleOmitYearValueI {

				monthVal := int(t.Month())
				dayVal := t.Day()

				contact.BirthdayMonth = utils.IntPtr(monthVal)
				contact.BirthdayDay = utils.IntPtr(dayVal)
			} else if !t.IsZero() {
				contact.Birthday = &t
			}
		}
	}

	// Anniversary
	if anniv := card.Get(vcard.FieldAnniversary); anniv != nil {
		anniversary := anniv.Value
		if strings.HasPrefix(anniversary, "--") {
			// Partial date format: --MMDD
			if len(anniversary) == 6 {
				month := utils.ParseIntPtr(anniversary[2:4])
				day := utils.ParseIntPtr(anniversary[4:6])
				contact.AnniversaryMonth = month
				contact.AnniversaryDay = day
			}
		} else {
			// Full date format: YYYYMMDD or YYYY-MM-DD
			t, _ := time.Parse("20060102", anniversary)
			if t.IsZero() {
				t, _ = time.Parse("2006-01-02", anniversary)
			}
			if !t.IsZero() {
				contact.Anniversary = &t
			}
		}
	}

	// Emails
	for _, field := range card[vcard.FieldEmail] {
		email := models.Email{
			Email:     field.Value,
			IsPrimary: field.Params.Get(vcard.ParamPreferred) == "1",
		}

		var labelToUse string

		// check first for a custom group
		if label := extractCustomLabel(card, field.Group); label != "" {
			labelToUse = label
		}

		// if none, look at the standard types
		if labelToUse == "" {
			for _, t := range field.Params.Types() {
				t = strings.ToLower(t)
				if t == "pref" {
					email.IsPrimary = true
					continue
				}
				// Take the first non-pref type we find (e.g., "work", "home")
				if labelToUse == "" {
					labelToUse = t
				}
			}
		}

		// fallback to cell
		if labelToUse == "" {
			labelToUse = "home"
		}

		key := getLabelKey("email", labelToUse)
		if id, ok := revMap[key]; ok {
			email.Type = id
		} else {
			// Here you could choose to auto-create the label in the DB
			// Or default to 'other'
			email.Type = revMap[getLabelKey("email", "home")]
		}

		contact.Emails = append(contact.Emails, email)
	}

	// Phones
	for _, field := range card[vcard.FieldTelephone] {
		phone := models.Phone{
			Phone:     field.Value,
			IsPrimary: field.Params.Get(vcard.ParamPreferred) == "1",
		}

		var labelToUse string

		// check first for a custom group
		if label := extractCustomLabel(card, field.Group); label != "" {
			labelToUse = label
		}

		// if none, look at the standard types
		if labelToUse == "" {
			for _, t := range field.Params.Types() {
				t = strings.ToLower(t)
				if t == "pref" {
					phone.IsPrimary = true
					continue
				}
				// Take the first non-pref type we find (e.g., "work", "home")
				if labelToUse == "" {
					labelToUse = t
				}
			}
		}

		// fallback to cell
		if labelToUse == "" {
			labelToUse = "cell"
		}

		key := getLabelKey("phone", labelToUse)
		if id, ok := revMap[key]; ok {
			phone.Type = id
		} else {
			// Here you could choose to auto-create the label in the DB
			// Or default to 'other'
			phone.Type = revMap[getLabelKey("phone", "cell")]
		}

		contact.Phones = append(contact.Phones, phone)
	}

	// Addresses
	for _, addr := range card.Addresses() {
		address := models.Address{
			Street:         addr.StreetAddress,
			ExtendedStreet: addr.ExtendedAddress,
			City:           addr.Locality,
			State:          addr.Region,
			PostalCode:     addr.PostalCode,
			Country:        addr.Country,
			IsPrimary:      addr.Field.Params.Get(vcard.ParamPreferred) == "1",
		}

		var labelToUse string

		// check first for a custom group
		if label := extractCustomLabel(card, addr.Field.Group); label != "" {
			labelToUse = label
		}

		// if none, look at the standard types
		if labelToUse == "" {
			for _, t := range addr.Field.Params.Types() {
				t = strings.ToLower(t)
				if t == "pref" {
					address.IsPrimary = true
					continue
				}
				// Take the first non-pref type we find (e.g., "work", "home")
				if labelToUse == "" {
					labelToUse = t
				}
			}
		}

		// fallback to cell
		if labelToUse == "" {
			labelToUse = "home"
		}

		key := getLabelKey("address", labelToUse)
		if id, ok := revMap[key]; ok {
			address.Type = id
		} else {
			// Here you could choose to auto-create the label in the DB
			// Or default to 'other'
			address.Type = revMap[getLabelKey("address", "home")]
		}
		contact.Addresses = append(contact.Addresses, address)
	}

	// Organization
	if org := card.Get(vcard.FieldOrganization); org != nil {
		organization := models.Organization{
			IsPrimary: true,
		}

		if phoneticOrg := card.Get(XPhoneticOrgField); phoneticOrg != nil {
			organization.PhoneticName = phoneticOrg.Value
		}

		parts := strings.Split(org.Value, ";")
		if len(parts) > 0 && parts[0] != "" {
			organization.Name = parts[0] // Company
		}
		if len(parts) > 1 && parts[1] != "" {
			organization.Department = parts[1] // Department
		}

		if title := card.Get(vcard.FieldTitle); title != nil {
			organization.Title = title.Value
		}
		contact.Organizations = append(contact.Organizations, organization)
	}

	// URLs
	for _, field := range card[vcard.FieldURL] {
		url := models.URL{URL: field.Value}

		var labelToUse string

		// check first for a custom group
		if label := extractCustomLabel(card, field.Group); label != "" {
			labelToUse = label
		}

		// if none, look at the standard types
		if labelToUse == "" {
			for _, t := range field.Params.Types() {
				t = strings.ToLower(t)
				if t == "pref" {
					url.IsPrimary = true
					continue
				}
				// Take the first non-pref type we find (e.g., "work", "home")
				if labelToUse == "" {
					labelToUse = t
				}
			}
		}

		// fallback to cell
		if labelToUse == "" {
			labelToUse = "home"
		}

		key := getLabelKey("url", labelToUse)
		if id, ok := revMap[key]; ok {
			url.Type = id
		} else {
			// Here you could choose to auto-create the label in the DB
			// Or default to 'other'
			url.Type = revMap[getLabelKey("url", "home")]
		}

		contact.URLs = append(contact.URLs, url)
	}

	// Notes
	if note := card.Get(vcard.FieldNote); note != nil {
		contact.Notes = note.Value
	}

	// Photo/Avatar
	if photo := card.Get(vcard.FieldPhoto); photo != nil {
		parts, err := parseVCardPhotoProperty(photo)
		if err != nil {
			fmt.Println("Error parsing data URL:", err)
			return nil, fmt.Errorf("failed to process photo property: %w", err)
		}
		contact.AvatarBase64 = parts.Base64Data
		contact.AvatarMimeType = parts.MimeType
	}

	// X-SOCIALPROFILE (iOS) -> convert to URL directly
	for _, field := range card[XSocialProfileField] {
		url := models.URL{URL: field.Value}

		var labelToUse string

		// check first for a custom group
		if label := extractCustomLabel(card, field.Group); label != "" {
			labelToUse = label
		}

		// if none, look at the standard types
		if labelToUse == "" {
			for _, t := range field.Params.Types() {
				t = strings.ToLower(t)
				if t == "pref" {
					url.IsPrimary = true
					continue
				}
				// Take the first non-pref type we find (e.g., "work", "home")
				if labelToUse == "" {
					labelToUse = t
				}
			}
		}

		// fallback to cell
		if labelToUse == "" {
			labelToUse = "profile"
		}

		key := getLabelKey("phone", labelToUse)
		if id, ok := revMap[key]; ok {
			url.Type = id
		} else {
			// Here you could choose to auto-create the label in the DB
			// Or default to 'other'
			url.Type = revMap[getLabelKey("url", "profile")]
		}

		contact.URLs = append(contact.URLs, url)
	}

	// X-ABLABELS and X-ABRELATEDNAME and X-ABDATE

	fieldsOfInterest := map[string]struct{}{
		XLabelField:        {},
		XDateField:         {},
		XRelatedNamesField: {},
	}

	contactLookup := make(map[string]*models.Contact)
	for _, c := range allContacts {
		contactLookup[strings.ToLower(c.FullName)] = c
	}

	relTypeLookup := make(map[string]*models.RelationshipType)
	for i := range allRelationshipTypes {
		rt := &allRelationshipTypes[i]
		relTypeLookup[strings.ToLower(rt.Name)] = rt
	}

	labelMap := make(map[string]string)
	relationshipMap := make(map[string]models.Relationship)
	otherDateMap := make(map[string]models.OtherDate)

	for fieldName, fieldSlice := range card {
		if _, ok := fieldsOfInterest[fieldName]; !ok {
			continue
		}

		for _, field := range fieldSlice {
			if !strings.HasPrefix(field.Group, "item") {
				continue
			}

			groupKey := field.Group

			switch fieldName {
			case XRelatedNamesField:
				rel := relationshipMap[groupKey]
				if rel.RelatedContact == nil {
					rel.RelatedContact = &models.Contact{}
				}
				if rel.RelationshipType == nil {
					rel.RelationshipType = &models.RelationshipType{}
				}
				rel.RelatedContact.FullName = strings.TrimSpace(field.Value)
				relationshipMap[groupKey] = rel

			case XDateField:
				od := otherDateMap[groupKey]
				dateVal := field.Value

				if strings.HasPrefix(dateVal, "1604-") && len(dateVal) == 10 { //Apple default of 1604 for blank
					// Apple format: 1604-MM-DD
					od.EventDateMonth = utils.ParseIntPtr(dateVal[5:7])
					od.EventDateDay = utils.ParseIntPtr(dateVal[8:10])
				} else if strings.HasPrefix(dateVal, "--") {
					// Standard No-Year: --MMDD or --MM-DD
					cleanDate := strings.ReplaceAll(dateVal, "-", "")
					if len(cleanDate) == 4 {
						od.EventDateMonth = utils.ParseIntPtr(cleanDate[0:2])
						od.EventDateDay = utils.ParseIntPtr(cleanDate[2:4])
					}
				} else {
					// Full Date: Try YYYY-MM-DD then YYYYMMDD
					t, _ := time.Parse("2006-01-02", dateVal)
					if t.IsZero() {
						t, _ = time.Parse("20060102", dateVal)
					}
					if !t.IsZero() {
						od.EventDate = &t
					}
				}
				otherDateMap[groupKey] = od

			case XLabelField:

				labelText := field.Value

				if after, ok := strings.CutPrefix(labelText, "_$!<"); ok {
					labelText = after
					labelText = strings.TrimSuffix(labelText, ">!$_")

					// Apple proprietary -> Format CamelCase to Space
					var schemaName strings.Builder
					for i, r := range labelText {
						if i > 0 && unicode.IsUpper(r) {
							schemaName.WriteString(" ")
						}
						schemaName.WriteString(string(r))
					}
					labelText = schemaName.String()
				}

				labelMap[groupKey] = labelText
			}
		}
	}

	// 5a. Finalize Other Dates
	for groupKey, otherDate := range otherDateMap {
		label := labelMap[groupKey]
		if label == "" {
			continue
		} // Ignore dates without labels (event name)

		otherDate.EventName = label
		hasFullDate := otherDate.EventDate != nil && !otherDate.EventDate.IsZero()
		hasPartialDate := otherDate.EventDateMonth != nil && otherDate.EventDateDay != nil

		if hasFullDate || hasPartialDate {
			if strings.EqualFold(otherDate.EventName, "Anniversary") {
				if hasFullDate {
					contact.Anniversary = otherDate.EventDate
				} else {
					contact.AnniversaryMonth = otherDate.EventDateMonth
					contact.AnniversaryDay = otherDate.EventDateDay
				}
			} else {
				contact.OtherDates = append(contact.OtherDates, otherDate)
			}
		}
	}

	// 5b. Finalize Relationships
	for groupKey, relationship := range relationshipMap {
		label := labelMap[groupKey]
		if label == "" {
			continue
		} // Ignore relationships without labels (relationship type)

		relationship.RelationshipType.Name = label
		if relationship.RelatedContact != nil && relationship.RelatedContact.FullName != "" {
			searchName := strings.ToLower(relationship.RelatedContact.FullName)
			searchLabel := strings.ToLower(label)

			// Search for matches
			matchedContact, contactFound := contactLookup[searchName]
			matchedRelType, relTypeFound := relTypeLookup[searchLabel]

			if contactFound && relTypeFound {
				// SUCCESS: We identified real KindredCard db values
				// Set the actual DB-backed objects
				relationship.RelatedContact = matchedContact
				relationship.RelationshipType = matchedRelType

				// Add to the main Relationships slice (foreign key relationship)
				contact.Relationships = append(contact.Relationships, relationship)
			} else {
				// FAIL: Insert as other_relationship (plain text storage)
				otherRelationship := models.OtherRelationship{
					RelatedContactName: relationship.RelatedContact.FullName,
					RelationshipName:   label,
				}
				contact.OtherRelationships = append(contact.OtherRelationships, otherRelationship)
			}
		}
	}

	return contact, nil
}

// VCardToContactShell converts a vCard to a limited Contact model: UID, FullName, and Gender
func VCardToContactShell(card vcard.Card) (*models.Contact, error) {
	uid := ""
	if field := card.Get(vcard.FieldUID); field != nil && field.Value != "" {
		uid = field.Value
	} else {
		uid = uuid.New().String() // generate a new UUID
	}

	contact := &models.Contact{
		UID: uid,
	}

	// Name fields
	if fn := card.Get(vcard.FieldFormattedName); fn != nil {
		contact.FullName = fn.Value
	}

	if n := card.Name(); n != nil {
		contact.GivenName = n.GivenName
		contact.FamilyName = n.FamilyName
		contact.MiddleName = n.AdditionalName
		contact.Prefix = n.HonorificPrefix
		contact.Suffix = n.HonorificSuffix
	}

	// If FullName is empty, generate it
	if contact.FullName == "" {
		contact.FullName = contact.GenerateFullName()
	}

	// Gender
	if gender := card.Get(vcard.FieldGender); gender != nil {
		contact.Gender = gender.Value
	}

	return contact, nil
}
