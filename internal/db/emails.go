package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactEmail(userID int, body models.Email) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactEmail(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Email:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO emails (contact_id, email, label_type_id, is_primary) VALUES ($1, $2, $3, $4) RETURNING id",
		body.ContactID, body.Email, body.Type, body.IsPrimary,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No emails inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating email: %v", err)
		return 0, fmt.Errorf("failed to create email: %w", err)
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

func (d *Database) UpdateContactEmail(userID int, body models.EmailJSONPatch) ([]models.Email, error) {
	logger.Debug("[DATABASE] Begin UpdateContactEmail(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of EmailJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.Email != nil {
		columns = append(columns, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *body.Email)
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

	// If nothing was sent to update, just return the current emails
	if len(columns) == 0 {
		return d.getEmails(*body.ContactID) // Helper to get contactID first if needed
	}

	query := fmt.Sprintf(`
        UPDATE emails 
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
			return nil, fmt.Errorf("email record not found or unauthorized")
		}
		logger.Error("Error patching email: %v", err)
		return nil, fmt.Errorf("failed to patch email: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.getEmails(contactID)
}

func (d *Database) DeleteContactEmail(userID int, contactID int, emailID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactEmail(userID:%d, contactID:%d, emailID:%d)", userID, contactID, emailID)

	_, err := d.db.Exec("DELETE FROM emails WHERE id = $1 AND contact_id = $2", emailID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Email: %v", err)
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
