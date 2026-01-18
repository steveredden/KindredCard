/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package db

import (
	"database/sql"
	"fmt"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

// GetUpcomingEventsByDays gets events in the next N days (1-14)
func (d *Database) GetUpcomingEventsByDays(userID int, days int) ([]models.UpcomingEvent, error) {
	logger.Debug("[DATABASE] Begin GetUpcomingEventsByDays(userID:%d, days:%d)", userID, days)

	query := `
	WITH upcoming_dates AS (
		SELECT 
			CURRENT_DATE + n * INTERVAL '1 day' as target_date,
			EXTRACT(MONTH FROM CURRENT_DATE + n * INTERVAL '1 day')::integer as target_month,
			EXTRACT(DAY FROM CURRENT_DATE + n * INTERVAL '1 day')::integer as target_day,
			n as days_offset
		FROM generate_series(0, $1) as n
	),
	birthdays AS (
		-- Birthdays with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			c.birthday as event_date,
			ud.target_date as this_year_date,
			ud.days_offset as days_until,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.birthday)::integer as age_years
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday IS NOT NULL
			AND EXTRACT(MONTH FROM c.birthday) = ud.target_month
			AND EXTRACT(DAY FROM c.birthday) = ud.target_day
		
		UNION ALL
		
		-- Birthdays with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			NULL as event_date,
			ud.target_date as this_year_date,
			ud.days_offset as days_until,
			NULL::integer as age_years
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday_month IS NOT NULL
			AND c.birthday_day IS NOT NULL
			AND c.birthday_month = ud.target_month
			AND c.birthday_day = ud.target_day
	),
	anniversaries AS (
		-- Anniversaries with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			c.anniversary as event_date,
			ud.target_date as this_year_date,
			ud.days_offset as days_until,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.anniversary)::integer as age_years
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary IS NOT NULL
			AND EXTRACT(MONTH FROM c.anniversary) = ud.target_month
			AND EXTRACT(DAY FROM c.anniversary) = ud.target_day
		
		UNION ALL
		
		-- Anniversaries with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			NULL as event_date,
			ud.target_date as this_year_date,
			ud.days_offset as days_until,
			NULL::integer as age_years
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary_month IS NOT NULL
			AND c.anniversary_day IS NOT NULL
			AND c.anniversary_month = ud.target_month
			AND c.anniversary_day = ud.target_day
	),
	other_events AS (
        -- Other events with full dates
        SELECT 
            c.id as contact_id,
            c.full_name,
            od.event_name as event_type,
            od.event_date,
            ud.target_date as this_year_date,
            ud.days_offset as days_until,
            EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM od.event_date)::integer as age_years
        FROM other_dates od
        JOIN contacts c ON od.contact_id = c.id
        CROSS JOIN upcoming_dates ud
        WHERE c.user_id = $2
			AND c.deleted_at IS NULL
            AND od.event_date IS NOT NULL
            AND EXTRACT(MONTH FROM od.event_date) = ud.target_month
            AND EXTRACT(DAY FROM od.event_date) = ud.target_day
        
        UNION ALL
        
        -- Other events with partial dates (month/day only)
        SELECT 
            c.id as contact_id,
            c.full_name,
            od.event_name as event_type,
            NULL as event_date,
            ud.target_date as this_year_date,
            ud.days_offset as days_until,
            NULL::integer as age_years
        FROM other_dates od
        JOIN contacts c ON od.contact_id = c.id
        CROSS JOIN upcoming_dates ud
        WHERE c.user_id = $2
			AND c.deleted_at IS NULL
            AND od.event_date_month IS NOT NULL
            AND od.event_date_day IS NOT NULL
            AND od.event_date_month = ud.target_month
            AND od.event_date_day = ud.target_day
    )
	SELECT 
		contact_id,
		full_name,
		event_type,
		event_date,
		this_year_date,
		days_until,
		age_years,
		CASE 
			WHEN days_until = 0 THEN 'Today'
			WHEN days_until = 1 THEN 'Tomorrow'
			ELSE days_until || ' days'
		END as time_description
	FROM (
		SELECT * FROM birthdays
		UNION ALL
		SELECT * FROM anniversaries
		UNION ALL
		SELECT * FROM other_events
	) all_events
	ORDER BY days_until, full_name, event_type
	`

	rows, err := d.db.Query(query, days, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting events: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var events []models.UpcomingEvent
	for rows.Next() {
		var event models.UpcomingEvent
		var eventDate sql.NullTime
		var ageYears sql.NullInt64

		err := rows.Scan(
			&event.ContactID,
			&event.FullName,
			&event.EventType,
			&eventDate,
			&event.ThisYearDate,
			&event.DaysUntil,
			&ageYears,
			&event.TimeDescription,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning events: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Convert nullable fields
		if eventDate.Valid {
			event.EventDate = &eventDate.Time
		}
		if ageYears.Valid {
			age := int(ageYears.Int64)
			event.AgeOrYears = &age
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		logger.Error("[DATABASE] Error for events: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

// GetUpcomingEventsByMonths gets events in the next N months (1-6)
func (d *Database) GetUpcomingEventsByMonths(userID int, months int) ([]models.UpcomingEvent, error) {
	logger.Debug("[DATABASE] Begin GetUpcomingEventsByMonths(userID:%d, months:%d)", userID, months)

	query := `
	WITH upcoming_months AS (
		SELECT 
			EXTRACT(MONTH FROM CURRENT_DATE + n * INTERVAL '1 month')::integer as target_month,
			EXTRACT(YEAR FROM CURRENT_DATE + n * INTERVAL '1 month')::integer as target_year,
			n as month_offset
		FROM generate_series(0, $1 - 1) as n
	),
	birthdays AS (
		-- Birthdays with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			c.birthday as event_date,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM c.birthday)::integer,
				EXTRACT(DAY FROM c.birthday)::integer
			) as this_year_date,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.birthday)::integer as age_years,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM c.birthday)::integer,
				EXTRACT(DAY FROM c.birthday)::integer
			) - CURRENT_DATE as days_until
		FROM contacts c
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday IS NOT NULL
			AND EXTRACT(MONTH FROM c.birthday)::integer = um.target_month
		
		UNION ALL
		
		-- Birthdays with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			NULL as event_date,
			MAKE_DATE(um.target_year, c.birthday_month, c.birthday_day) as this_year_date,
			NULL::integer as age_years,
			MAKE_DATE(um.target_year, c.birthday_month, c.birthday_day) - CURRENT_DATE as days_until
		FROM contacts c
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday_month IS NOT NULL
			AND c.birthday_day IS NOT NULL
			AND c.birthday_month = um.target_month
	),
	anniversaries AS (
		-- Anniversaries with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			c.anniversary as event_date,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM c.anniversary)::integer,
				EXTRACT(DAY FROM c.anniversary)::integer
			) as this_year_date,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.anniversary)::integer as age_years,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM c.anniversary)::integer,
				EXTRACT(DAY FROM c.anniversary)::integer
			) - CURRENT_DATE as days_until
		FROM contacts c
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary IS NOT NULL
			AND EXTRACT(MONTH FROM c.anniversary)::integer = um.target_month
		
		UNION ALL
		
		-- Anniversaries with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			NULL as event_date,
			MAKE_DATE(um.target_year, c.anniversary_month, c.anniversary_day) as this_year_date,
			NULL::integer as age_years,
			MAKE_DATE(um.target_year, c.anniversary_month, c.anniversary_day) - CURRENT_DATE as days_until
		FROM contacts c
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary_month IS NOT NULL
			AND c.anniversary_day IS NOT NULL
			AND c.anniversary_month = um.target_month
	),
	other_events AS (
		-- Other events with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			od.event_name as event_type,
			od.event_date,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM od.event_date)::integer,
				EXTRACT(DAY FROM od.event_date)::integer
			) as this_year_date,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM od.event_date)::integer as age_years,
			MAKE_DATE(
				um.target_year,
				EXTRACT(MONTH FROM od.event_date)::integer,
				EXTRACT(DAY FROM od.event_date)::integer
			) - CURRENT_DATE as days_until
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND od.event_date IS NOT NULL
			AND EXTRACT(MONTH FROM od.event_date)::integer = um.target_month
		
		UNION ALL
		
		-- Other events with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			od.event_name as event_type,
			NULL as event_date,
			MAKE_DATE(um.target_year, od.event_date_month, od.event_date_day) as this_year_date,
			NULL::integer as age_years,
			MAKE_DATE(um.target_year, od.event_date_month, od.event_date_day) - CURRENT_DATE as days_until
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN upcoming_months um
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND od.event_date_month IS NOT NULL
			AND od.event_date_day IS NOT NULL
			AND od.event_date_month = um.target_month
	)
	SELECT 
		contact_id,
		full_name,
		event_type,
		event_date,
		this_year_date,
		days_until,
		age_years,
		CASE 
			WHEN days_until = 0 THEN 'Today'
			WHEN days_until = 1 THEN 'Tomorrow'
			WHEN days_until < 7 THEN days_until || ' days'
			WHEN days_until < 14 THEN (days_until / 7) || ' week'
			ELSE (days_until / 7) || ' weeks'
		END as time_description
	FROM (
		SELECT * FROM birthdays
		UNION ALL
		SELECT * FROM anniversaries
		UNION ALL
		SELECT * FROM other_events
	) all_events
	WHERE days_until >= 0
	ORDER BY this_year_date, full_name, event_type
	`

	rows, err := d.db.Query(query, months, userID)
	if err != nil {
		logger.Error("[DATABASE] Error selecting events: %v", err)
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var events []models.UpcomingEvent
	for rows.Next() {
		var event models.UpcomingEvent
		var eventDate sql.NullTime
		var ageYears sql.NullInt64

		err := rows.Scan(
			&event.ContactID,
			&event.FullName,
			&event.EventType,
			&eventDate,
			&event.ThisYearDate,
			&event.DaysUntil,
			&ageYears,
			&event.TimeDescription,
		)
		if err != nil {
			logger.Error("[DATABASE] Error scanning events: %v", err)
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Convert nullable fields
		if eventDate.Valid {
			event.EventDate = &eventDate.Time
		}
		if ageYears.Valid {
			age := int(ageYears.Int64)
			event.AgeOrYears = &age
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		logger.Error("[DATABASE] Error for events: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

// GetUpcomingEventsCount gets the count of upcoming events
// Useful for dashboard badges/indicators
func (d *Database) GetUpcomingEventsCount(userID int, days int) (int, error) {
	logger.Debug("[DATABASE] Begin GetUpcomingEventsCount(userID:%d, days:%d)", userID, days)

	query := `
	WITH upcoming_dates AS (
		SELECT 
			EXTRACT(MONTH FROM CURRENT_DATE + n * INTERVAL '1 day')::integer as target_month,
			EXTRACT(DAY FROM CURRENT_DATE + n * INTERVAL '1 day')::integer as target_day
		FROM generate_series(0, $1) as n
	)
	SELECT COUNT(*) FROM (
		-- Birthdays with full dates
		SELECT 1
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday IS NOT NULL
			AND EXTRACT(MONTH FROM c.birthday) = ud.target_month
			AND EXTRACT(DAY FROM c.birthday) = ud.target_day
		
		UNION ALL
		
		-- Birthdays with partial dates
		SELECT 1
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday_month = ud.target_month
			AND c.birthday_day = ud.target_day
		
		UNION ALL
		
		-- Anniversaries with full dates
		SELECT 1
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary IS NOT NULL
			AND EXTRACT(MONTH FROM c.anniversary) = ud.target_month
			AND EXTRACT(DAY FROM c.anniversary) = ud.target_day
		
		UNION ALL
		
		-- Anniversaries with partial dates
		SELECT 1
		FROM contacts c
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary_month = ud.target_month
			AND c.anniversary_day = ud.target_day
		
		UNION ALL
		
		-- Other events with full dates
		SELECT 1
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND od.event_date IS NOT NULL
			AND EXTRACT(MONTH FROM od.event_date) = ud.target_month
			AND EXTRACT(DAY FROM od.event_date) = ud.target_day
		
		UNION ALL
		
		-- Other events with partial dates
		SELECT 1
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN upcoming_dates ud
		WHERE c.user_id = $2
			AND od.event_date_month = ud.target_month
			AND od.event_date_day = ud.target_day
	) all_events
	`

	var count int
	err := d.db.QueryRow(query, days, userID).Scan(&count)
	if err != nil {
		logger.Error("[DATABASE] Error selecting events: %v", err)
		return 0, fmt.Errorf("count query error: %w", err)
	}

	return count, nil
}

// GetRecentPastEventsByDays gets events that occurred in the past N days (lookback)
// Returns events with negative days_until values (e.g., -1 for yesterday, -7 for a week ago)
// Useful for "last chance" reminders or missed event notifications
func (d *Database) GetRecentPastEventsByDays(userID int, lookbackDays int) ([]models.UpcomingEvent, error) {
	logger.Debug("[DATABASE] Begin GetRecentPastEventsByDays(userID:%d, lookbackDays:%d)", userID, lookbackDays)

	query := `
	WITH past_dates AS (
		SELECT 
			CURRENT_DATE - n * INTERVAL '1 day' as target_date,
			EXTRACT(MONTH FROM CURRENT_DATE - n * INTERVAL '1 day')::integer as target_month,
			EXTRACT(DAY FROM CURRENT_DATE - n * INTERVAL '1 day')::integer as target_day,
			-n as days_offset
		FROM generate_series(1, $1) as n
	),
	birthdays AS (
		-- Birthdays with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			c.birthday as event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.birthday)::integer as age_years
		FROM contacts c
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
	        AND c.deleted_at IS NULL
			AND c.birthday IS NOT NULL
			AND EXTRACT(MONTH FROM c.birthday) = pd.target_month
			AND EXTRACT(DAY FROM c.birthday) = pd.target_day
		
		UNION ALL
		
		-- Birthdays with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'birthday' as event_type,
			NULL as event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			NULL as age_years
		FROM contacts c
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.birthday_month IS NOT NULL
			AND c.birthday_day IS NOT NULL
			AND c.birthday_month = pd.target_month
			AND c.birthday_day = pd.target_day
	),
	anniversaries AS (
		-- Anniversaries with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			c.anniversary as event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM c.anniversary)::integer as age_years
		FROM contacts c
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary IS NOT NULL
			AND EXTRACT(MONTH FROM c.anniversary) = pd.target_month
			AND EXTRACT(DAY FROM c.anniversary) = pd.target_day
		
		UNION ALL
		
		-- Anniversaries with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			'anniversary' as event_type,
			NULL as event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			NULL as age_years
		FROM contacts c
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND c.anniversary_month IS NOT NULL
			AND c.anniversary_day IS NOT NULL
			AND c.anniversary_month = pd.target_month
			AND c.anniversary_day = pd.target_day
	),
	other_events AS (
		-- Other events with full dates
		SELECT 
			c.id as contact_id,
			c.full_name,
			od.event_name as event_type,
			od.event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			EXTRACT(YEAR FROM CURRENT_DATE)::integer - EXTRACT(YEAR FROM od.event_date)::integer as age_years
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND od.event_date IS NOT NULL
			AND EXTRACT(MONTH FROM od.event_date) = pd.target_month
			AND EXTRACT(DAY FROM od.event_date) = pd.target_day
		
		UNION ALL
		
		-- Other events with partial dates (month/day only)
		SELECT 
			c.id as contact_id,
			c.full_name,
			od.event_name as event_type,
			NULL as event_date,
			pd.target_date as this_year_date,
			pd.days_offset as days_until,
			NULL::integer as age_years
		FROM other_dates od
		JOIN contacts c ON od.contact_id = c.id
		CROSS JOIN past_dates pd
		WHERE c.user_id = $2
			AND c.deleted_at IS NULL
			AND od.event_date_month IS NOT NULL
			AND od.event_date_day IS NOT NULL
			AND od.event_date_month = pd.target_month
			AND od.event_date_day = pd.target_day
	)
	SELECT 
		contact_id,
		full_name,
		event_type,
		event_date,
		this_year_date,
		days_until,
		age_years,
		CASE 
			WHEN days_until = -1 THEN 'Yesterday'
			WHEN days_until < -1 THEN ABS(days_until) || ' days ago'
			ELSE 'Today'
		END as time_description
	FROM (
		SELECT * FROM birthdays
		UNION ALL
		SELECT * FROM anniversaries
		UNION ALL
		SELECT * FROM other_events
	) all_events
	ORDER BY days_until DESC, full_name, event_type
	`

	rows, err := d.db.Query(query, lookbackDays, userID)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var events []models.UpcomingEvent
	for rows.Next() {
		var event models.UpcomingEvent
		var eventDate sql.NullTime
		var ageYears sql.NullInt64

		err := rows.Scan(
			&event.ContactID,
			&event.FullName,
			&event.EventType,
			&eventDate,
			&event.ThisYearDate,
			&event.DaysUntil,
			&ageYears,
			&event.TimeDescription,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Convert nullable fields
		if eventDate.Valid {
			event.EventDate = &eventDate.Time
		}
		if ageYears.Valid {
			age := int(ageYears.Int64)
			event.AgeOrYears = &age
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

// GetLastWeeksPastEvents is a convenience function for getting events from the past week
func (d *Database) GetLastWeeksPastEvents(userID int) ([]models.UpcomingEvent, error) {
	return d.GetRecentPastEventsByDays(userID, 7)
}

// GetTodaysEvents is a convenience function for getting today's events
func (d *Database) GetTodaysEvents(userID int) ([]models.UpcomingEvent, error) {
	return d.GetUpcomingEventsByDays(userID, 0)
}

// GetThisWeeksEvents is a convenience function for getting this week's events
func (d *Database) GetThisWeeksEvents(userID int) ([]models.UpcomingEvent, error) {
	// Get events up to 7 days out, but use the days function with max 3
	// For longer periods, use the months function
	return d.GetUpcomingEventsByDays(userID, 7)
}
