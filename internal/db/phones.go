package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactPhone(userID int, body models.Phone) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactPhone(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Phone:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO phones (contact_id, phone, label_type_id, is_primary) VALUES ($1, $2, $3, $4) RETURNING id",
		body.ContactID, body.Phone, body.Type, body.IsPrimary,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No phones inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating phone: %v", err)
		return 0, fmt.Errorf("failed to create phone: %w", err)
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

func (d *Database) UpdateContactPhone(userID int, body models.PhoneJSONPatch) ([]models.Phone, error) {
	logger.Debug("[DATABASE] Begin UpdateContactPhone(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of PhoneJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.Phone != nil {
		columns = append(columns, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *body.Phone)
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
		return d.getPhones(*body.ContactID) // Helper to get contactID first if needed
	}

	columns = append(columns, "last_formatted_at = NOW()")

	query := fmt.Sprintf(`
        UPDATE phones 
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

	return d.getPhones(contactID)
}

func (d *Database) DeleteContactPhone(userID int, contactID int, phoneID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactPhone(userID:%d, contactID:%d, phoneID:%d)", userID, contactID, phoneID)

	_, err := d.db.Exec("DELETE FROM phones WHERE id = $1 AND contact_id = $2", phoneID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Phone: %v", err)
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
