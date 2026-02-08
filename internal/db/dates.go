package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactOtherDate(userID int, body models.ContactDateJSONPatch) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactOtherDate(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Other Date:")
		utils.Dump(body)
	}

	var newID int
	var query string
	args := []interface{}{}
	args = append(args, body.ContactID)
	args = append(args, body.DateType)

	if body.Date != nil {
		query = "INSERT INTO other_dates (contact_id, event_name, event_date) VALUES ($1, $2, $3) RETURNING id"
		args = append(args, body.Date)
	} else if body.DateMonth != nil && body.DateDay != nil {
		query = "INSERT INTO other_dates (contact_id, event_name, event_date_month, event_date_day) VALUES ($1, $2, $3, $4) RETURNING id"
		args = append(args, body.DateMonth)
		args = append(args, body.DateDay)
	}

	err := d.db.QueryRow(query, args...).Scan(&newID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No other dates inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating other date: %v", err)
		return 0, fmt.Errorf("failed to create other date: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return newID, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(body.ContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return newID, nil
}

// UpdateContactDate specifically updates other_dates on a contact
func (d *Database) UpdateContactOtherDate(userID int, body models.ContactDateJSONPatch) error {
	logger.Debug("[DATABASE] Begin UpdateContactOtherDate(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of ContactDateJSONPatch:")
		utils.Dump(body)
	}

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if body.DateType != "" {
		updates = append(updates, fmt.Sprintf("event_name = $%d", argIndex))
		args = append(args, body.DateType)
		argIndex++
	}

	if body.Date != nil {
		updates = append(updates, fmt.Sprintf("event_date = $%d", argIndex))
		args = append(args, body.Date)
		argIndex++
	} else if body.DateMonth != nil && body.DateDay != nil {
		updates = append(updates, fmt.Sprintf("event_date_month = $%d", argIndex))
		args = append(args, body.DateMonth)
		argIndex++

		updates = append(updates, fmt.Sprintf("event_date_day = $%d", argIndex))
		args = append(args, body.DateDay)
		argIndex++
	}

	// Add WHERE clause parameters
	//    WHERE id = $%d
	args = append(args, body.ContactID)

	// Build and execute query
	query := fmt.Sprintf(`
		UPDATE other_dates
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex+1)

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return err
	}

	// Increment sync token so CardDAV clients pull the change
	newToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err == nil {
		_ = d.bumpContactSyncToken(body.ContactID, newToken)
	}

	return nil
}

// UpdateContactDate specifically updates formal dates (birthday / anniversary) on a contact
func (d *Database) UpdateContactDate(userID int, body models.ContactDateJSONPatch) error {
	logger.Debug("[DATABASE] Begin UpdateContactDate(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of ContactDateJSONPatch:")
		utils.Dump(body)
	}

	tableName := ""
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	switch body.DateType {
	case "anniversary", "birthday":
		tableName = "contacts"

		// 1. If we have a full date (ISO string), we update the main column
		// AND we MUST clear the partial month/day columns to keep data clean.
		if body.Date != nil {
			updates = append(updates, fmt.Sprintf("%s = $%d", body.DateType, argIndex))
			args = append(args, body.Date)
			argIndex++

			updates = append(updates, fmt.Sprintf("%s_month = NULL", body.DateType))
			updates = append(updates, fmt.Sprintf("%s_day = NULL", body.DateType))
		} else {
			// 2. If body.Date is nil, it means either we are clearing EVERYTHING,
			// or we are using partial dates (Month/Day).

			// Update the full date column to NULL (clearing the YYYY-MM-DD version)
			updates = append(updates, fmt.Sprintf("%s = NULL", body.DateType))

			// Now handle the month/day pointers (will be nil if cleared)
			updates = append(updates, fmt.Sprintf("%s_month = $%d", body.DateType, argIndex))
			args = append(args, body.DateMonth) // driver handles nil as SQL NULL
			argIndex++

			updates = append(updates, fmt.Sprintf("%s_day = $%d", body.DateType, argIndex))
			args = append(args, body.DateDay) // driver handles nil as SQL NULL
			argIndex++
		}
	}

	// Add WHERE clause parameters
	//    WHERE id = $%d AND user_id = $%d
	args = append(args, body.ContactID, userID)

	// Build and execute query
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id = $%d AND user_id = $%d
	`, tableName, strings.Join(updates, ", "), argIndex, argIndex+1)

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return err
	}

	// Increment sync token so CardDAV clients pull the change
	newToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err == nil {
		_ = d.bumpContactSyncToken(body.ContactID, newToken)
	}

	return nil
}

func (d *Database) DeleteContactOtherDate(userID int, contactID int, otherDateID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactOtherDate(userID:%d, contactID:%d, emailID:%d)", userID, contactID, otherDateID)

	_, err := d.db.Exec("DELETE FROM other_dates WHERE id = $1 AND contact_id = $2", otherDateID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Other Date: %v", err)
		return err
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return nil
}
