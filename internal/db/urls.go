package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactURL(userID int, body models.URL) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactURL(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of URL:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO urls (contact_id, url, label_type_id) VALUES ($1, $2, $3) RETURNING id",
		body.ContactID, body.URL, body.Type,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No urls inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating url: %v", err)
		return 0, fmt.Errorf("failed to create url: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return body.ID, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(body.ContactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return body.ID, nil
}

func (d *Database) UpdateContactURL(userID int, body models.URLJSONPatch) ([]models.URL, error) {
	logger.Debug("[DATABASE] Begin UpdateContactURL(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of URLJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.URL != nil {
		columns = append(columns, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *body.URL)
		argIdx++
	}

	if body.Type != nil {
		columns = append(columns, fmt.Sprintf("label_type_id = $%d", argIdx))
		args = append(args, *body.Type)
		argIdx++
	}

	if body.IsPrimary != nil {
		columns = append(columns, fmt.Sprintf("is_primary = $%d", argIdx))
		args = append(args, *body.IsPrimary)
		argIdx++
	}

	// If nothing was sent to update, just return the current phones
	if len(columns) == 0 {
		return d.getURLs(*body.ContactID) // Helper to get contactID first if needed
	}

	query := fmt.Sprintf(`
        UPDATE urls 
        SET %s
        WHERE id = $%d 
        AND contact_id IN (SELECT id FROM contacts WHERE user_id = $%d)
        RETURNING contact_id`,
		strings.Join(columns, ", "),
		argIdx,
		argIdx+1,
	)

	args = append(args, body.ID, userID)

	var contactID int
	err := d.db.QueryRow(query, args...).Scan(&contactID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No rows patched: %v", err)
			return nil, fmt.Errorf("phone record not found or unauthorized")
		}
		logger.Error("Error patching phone: %v", err)
		return nil, fmt.Errorf("failed to patch phone: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.getURLs(contactID)
}

func (d *Database) DeleteContactURL(userID int, contactID int, urlID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactURL(userID:%d, contactID:%d, urlID:%d)", userID, contactID, urlID)

	_, err := d.db.Exec("DELETE FROM urls WHERE id = $1 AND contact_id = $2", urlID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting URL: %v", err)
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
