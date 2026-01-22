/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

// DiscordEmbed represents a Discord embed
type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordFooter      `json:"footer,omitempty"`
}

// DiscordEmbedField represents a field in a Discord embed
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordFooter represents embed footer
type DiscordFooter struct {
	Text string `json:"text"`
}

// DiscordWebhook represents the payload sent to Discord
type DiscordWebhook struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// SendDiscordNotification sends a notification to Discord
func SendDiscordNotification(webhookURL string, embeds []DiscordEmbed) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	webhook := DiscordWebhook{
		Embeds: embeds,
	}

	data, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("error marshaling webhook: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error posting to Discord: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// BuildTodayEventsEmbed creates a Discord embed for today's events
func BuildTodayEventsEmbed(events []models.UpcomingEvent, baseURL string) DiscordEmbed {

	title := fmt.Sprintf("🎉 Today's Events - %s", time.Now().Local().Format("Jan 2"))

	embed := DiscordEmbed{
		Title:     title,
		Color:     0x5865F2, // Discord blurple
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &DiscordFooter{
			Text: "KindredCard",
		},
	}

	if len(events) == 0 {
		embed.Description = "No birthdays or anniversaries today!"
		return embed
	}

	// Group by event type
	birthdays := []models.UpcomingEvent{}
	anniversaries := []models.UpcomingEvent{}
	others := []models.UpcomingEvent{}

	for _, event := range events {
		switch event.EventType {
		case "birthday":
			birthdays = append(birthdays, event)
		case "anniversary":
			anniversaries = append(anniversaries, event)
		default:
			others = append(others, event)
		}
	}

	// Add birthday field
	if len(birthdays) > 0 {
		var birthdayText string
		for _, b := range birthdays {
			nameHyperlink := makeHyperlink(b.FullName, fmt.Sprintf("%s/contacts/%d", baseURL, b.ContactID))
			if b.AgeOrYears != nil {
				ordinal := utils.Ordinal(*b.AgeOrYears)
				birthdayText += fmt.Sprintf("🎂 **%s** - %s birthday %s!\n", nameHyperlink, ordinal, b.TimeDescription)
			} else {
				birthdayText += fmt.Sprintf("🎂 **%s** - has a birthday %s!\n", nameHyperlink, b.TimeDescription)
			}
		}
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:  "Birthdays",
			Value: birthdayText,
		})
	}

	// Add anniversary field
	if len(anniversaries) > 0 {
		var anniversaryText string
		for _, a := range anniversaries {
			nameHyperlink := makeHyperlink(a.FullName, fmt.Sprintf("%s/contacts/%d", baseURL, a.ContactID))
			if a.AgeOrYears != nil {
				ordinal := utils.Ordinal(*a.AgeOrYears)
				anniversaryText += fmt.Sprintf("💍 **%s** - %s wedding anniversary %s!\n", nameHyperlink, ordinal, a.TimeDescription)
			} else {
				anniversaryText += fmt.Sprintf("💍 **%s** - has a wedding anniversary %s!\n", nameHyperlink, a.TimeDescription)
			}
		}
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:  "Anniversaries",
			Value: anniversaryText,
		})
	}

	// Add other dates field
	if len(others) > 0 {
		var otherText string
		for _, o := range others {
			nameHyperlink := makeHyperlink(o.FullName, fmt.Sprintf("%s/contacts/%d", baseURL, o.ContactID))
			if o.AgeOrYears != nil {
				ordinal := utils.Ordinal(*o.AgeOrYears)
				otherText += fmt.Sprintf("📅 **%s** - %s anniversary of %s %s!\n", nameHyperlink, ordinal, o.EventType, o.TimeDescription)
			} else {
				otherText += fmt.Sprintf("📅 **%s** - anniversary of %s %s!\n", nameHyperlink, o.EventType, o.TimeDescription)
			}
		}
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:  "Other Dates",
			Value: otherText,
		})
	}

	return embed
}

// SendTestNotification sends a test notification with dummy data
func SendTestNotification(webhookURL string, baseURL string) error {
	dummyAge1 := 30
	dummyAge2 := 2

	dummyEvents := []models.UpcomingEvent{
		{
			ContactID:       99999,
			FullName:        "John Doe",
			EventType:       "birthday",
			AgeOrYears:      &dummyAge1,
			TimeDescription: "Today",
		},
		{
			ContactID:       99998,
			FullName:        "Jane Smith",
			EventType:       "anniversary",
			AgeOrYears:      &dummyAge2,
			TimeDescription: "Tomorrow",
		},
		{
			ContactID:       99997,
			FullName:        "Jack Jones",
			EventType:       "Retirement",
			AgeOrYears:      &dummyAge2,
			TimeDescription: "in 3 days",
		},
	}

	embed := BuildTodayEventsEmbed(dummyEvents, baseURL)
	embed.Description = "This is a 🧪 test notification from KindredCard!"
	embed.Color = 0xFFA500 // Orange for test

	return SendDiscordNotification(webhookURL, []DiscordEmbed{embed})
}

func makeHyperlink(display string, url string) string {
	return fmt.Sprintf("[%s](%s)", display, url)
}
