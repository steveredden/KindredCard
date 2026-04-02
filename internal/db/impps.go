package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactIMPP(userID int, body models.IMPP) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactIMPP(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of IMPP:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO impps (contact_id, impp, label_type_id) VALUES ($1, $2, $3) RETURNING id",
		body.ContactID, body.IMPP, body.Type,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No impps inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating impp: %v", err)
		return 0, fmt.Errorf("failed to create impp: %w", err)
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

func (d *Database) UpdateContactIMPP(userID int, body models.IMPPJSONPatch) ([]models.IMPP, error) {
	logger.Debug("[DATABASE] Begin UpdateContactIMPP(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of IMPPJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.IMPP != nil {
		columns = append(columns, fmt.Sprintf("impp = $%d", argIdx))
		args = append(args, *body.IMPP)
		argIdx++
	}

	if body.Type != nil {
		columns = append(columns, fmt.Sprintf("label_type_id = $%d", argIdx))
		args = append(args, *body.Type)
		argIdx++
	}

	// If nothing was sent to update, just return the current impps
	if len(columns) == 0 {
		return d.getIMPPs(*body.ContactID) // Helper to get contactID first if needed
	}

	query := fmt.Sprintf(`
        UPDATE impps 
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
			return nil, fmt.Errorf("impp record not found or unauthorized")
		}
		logger.Error("Error patching impp: %v", err)
		return nil, fmt.Errorf("failed to patch impp: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.getIMPPs(contactID)
}

func (d *Database) DeleteContactIMPP(userID int, contactID int, imppID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactIMPP(userID:%d, contactID:%d, imppID:%d)", userID, contactID, imppID)

	_, err := d.db.Exec("DELETE FROM impps WHERE id = $1 AND contact_id = $2", imppID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting IMPP: %v", err)
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
