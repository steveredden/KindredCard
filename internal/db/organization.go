package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactOrganization(userID int, body models.Organization) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactOrganization(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Organization:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO organizations (contact_id, name, title, department) VALUES ($1, $2, $3, $4) RETURNING id",
		body.ContactID, body.Name, body.Title, body.Department,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No organizations inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating organization: %v", err)
		return 0, fmt.Errorf("failed to create organization: %w", err)
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

func (d *Database) UpdateContactOrganization(userID int, body models.OrganizationJSONPatch) ([]models.Organization, error) {
	logger.Debug("[DATABASE] Begin UpdateContactOrganization(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of OrganizationJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.Name != nil {
		columns = append(columns, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *body.Name)
		argIdx++
	}

	if body.Title != nil {
		columns = append(columns, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *body.Title)
		argIdx++
	}

	if body.Department != nil {
		columns = append(columns, fmt.Sprintf("department = $%d", argIdx))
		args = append(args, *body.Department)
		argIdx++
	}

	// If nothing was sent to update, just return the current organizations
	if len(columns) == 0 {
		return d.getOrganizations(*body.ContactID) // Helper to get contactID first if needed
	}

	query := fmt.Sprintf(`
        UPDATE organizations 
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
			return nil, fmt.Errorf("organization record not found or unauthorized")
		}
		logger.Error("Error patching organization: %v", err)
		return nil, fmt.Errorf("failed to patch organization: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.getOrganizations(contactID)
}

func (d *Database) DeleteContactOrganization(userID int, contactID int, organizationID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactOrganization(userID:%d, contactID:%d, organizationID:%d)", userID, contactID, organizationID)

	_, err := d.db.Exec("DELETE FROM organizations WHERE id = $1 AND contact_id = $2", organizationID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Organization: %v", err)
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
