/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package scheduler

import (
	"regexp"
	"time"

	"github.com/steveredden/KindredCard/internal/db"
	"github.com/steveredden/KindredCard/internal/discord"
	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/mailer"
	"github.com/steveredden/KindredCard/internal/models"
)

// Scheduler handles scheduled notification checks
type Scheduler struct {
	db       *db.Database
	ticker   *time.Ticker
	stopChan chan bool
	baseURL  string
}

// NewScheduler creates a new scheduler instance
func NewScheduler(db *db.Database, baseURL string) *Scheduler {
	return &Scheduler{
		db:       db,
		stopChan: make(chan bool),
		baseURL:  baseURL,
	}
}

// Start begins the scheduler with 1-minute interval checks
func (s *Scheduler) Start() {
	logger.Info("[SCHEDULER] Scheduler started - checking every minute")

	// Run immediately on startup
	s.globalCleanup(true)
	s.checkAndSendNotifications()

	// Then run every minute
	s.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.globalCleanup(false)
				s.checkAndSendNotifications()
			case <-s.stopChan:
				logger.Info("📅 Scheduler stopped")
				return
			}
		}
	}()
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.stopChan <- true
}

// checkAndSendNotifications checks all notification settings and sends due notifications
func (s *Scheduler) checkAndSendNotifications() {
	now := time.Now()
	currentTime := now.Format("15:04") // HH:MM format

	logger.Debug("[SCHEDULER] Checking notifications scheduled at %s", currentTime)

	// Get all enabled notification settings
	notifiers, err := s.db.GetAllNotificationSettings(true)
	if err != nil {
		logger.Error("[SCHEDULER] Error fetching notification settings: %v", err)
		return
	}

	if len(notifiers) == 0 {
		logger.Debug("[SCHEDULER] No enabled notification settings found")
		return
	}

	// Process each notifier setting
	for _, setting := range notifiers {
		// Check if it's time to send this notification
		if setting.NotificationTime == currentTime {
			logger.Info("[SCHEDULER] Time match for setting #%d at %s", setting.ID, currentTime)
			s.processNotificationSetting(setting)
		}
	}
}

// processNotificationSetting processes a single notification setting
func (s *Scheduler) processNotificationSetting(setting models.NotificationSetting) {
	logger.Debug("[SCHEDULER] Processing notification setting: %s [%d]", setting.Name, setting.ID)

	// Check if we've already sent this notification today
	if s.db.HasNotificationBeenSent(setting) {
		logger.Info("[SCHEDULER] Skipping - already sent notification %s [%d]", setting.Name, setting.ID)
		return
	}

	// Get upcoming events for this setting
	events, err := s.db.GetUpcomingEventsByDays(setting.UserID, setting.DaysLookAhead)
	if err != nil {
		logger.Error("[SCHEDULER] Error getting upcoming events: %v", err)
		return
	}

	var re *regexp.Regexp
	testRegex := false

	if setting.EventRegex != "" {
		pattern := "(?i)" + setting.EventRegex
		var err error //init so we can assign value, not :=
		re, err = regexp.Compile(pattern)
		if err != nil {
			logger.Error("Invalid regex pattern '%s': %v", pattern, err)
			testRegex = false
		} else {
			testRegex = true
		}
	}

	relevantEvents := []models.UpcomingEvent{}
	for _, event := range events {

		include := false

		if setting.IncludeBirthdays && event.EventType == "birthday" {
			include = true
		} else if setting.IncludeAnniversaries && event.EventType == "anniversary" {
			include = true
		} else if setting.IncludeEventDates {
			if testRegex {
				logger.Debug("[SCHEDULER] Evaluating Regex pattern: '%s' -> '%s'", setting.EventRegex, event.EventType)
				if re.MatchString(event.EventType) {
					include = true
				}
			} else {
				include = true
			}
		}

		if include {
			desc := event.TimeDescription
			if desc != "Today" && desc != "Tomorrow" {
				event.TimeDescription = "in " + desc
			}

			relevantEvents = append(relevantEvents, event)
		}
	}

	if len(relevantEvents) == 0 {
		logger.Info("[SCHEDULER] No upcoming events found for webhook %s [%d]", setting.Name, setting.ID)
		return
	}

	logger.Debug("[SCHEDULER] Found %d upcoming events", len(relevantEvents))

	switch setting.ProviderType {
	case "discord":
		if setting.WebhookURL == nil || *setting.WebhookURL == "" {
			logger.Warn("Discord notification enabled but no WebhookURL provided")
		}
		embed := discord.BuildTodayEventsEmbed(relevantEvents, s.baseURL)
		discord.SendDiscordNotification(*setting.WebhookURL, []discord.DiscordEmbed{embed})
	case "smtp":
		if setting.TargetAddress == nil || *setting.TargetAddress == "" {
			logger.Warn("SMTP notification enabled but no TargetAddress provided")
			return
		}
		body := mailer.BuildTodayEventsBody(relevantEvents, s.baseURL)
		mailer.SendEventNotification(*setting.TargetAddress, "KindredCard Event Summary", body.Body)
	}

	// Record that we sent this notification
	if err := s.db.RecordNotificationSettingSent(setting); err != nil {
		logger.Error("[SCHEDULER] Error recording notification: %v", err)
	}
}
