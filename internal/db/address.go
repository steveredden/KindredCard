package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

func (d *Database) CreateContactAddress(userID int, body models.Address) (int, error) {
	logger.Debug("[DATABASE] Begin CreateContactAddress(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of Address:")
		utils.Dump(body)
	}

	err := d.db.QueryRow(
		"INSERT INTO addresses (contact_id, street, extended_street, city, state, postal_code, country, label_type_id, is_primary) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		body.ContactID, body.Street, body.ExtendedStreet, body.City, body.State, body.PostalCode, body.Country, body.Type, body.IsPrimary,
	).Scan(&body.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("No addresses inserted: %v", err)
			return 0, fmt.Errorf("unauthorized")
		}
		logger.Error("Error creating address: %v", err)
		return 0, fmt.Errorf("failed to create address: %w", err)
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

func (d *Database) UpdateContactAddress(userID int, body models.AddressJSONPatch) ([]models.Address, error) {
	logger.Debug("[DATABASE] Begin UpdateContactAddress(userID:%d, body:--)", userID)

	if logger.GetLevel() == logger.TRACE {
		logger.Trace("[DATABSE] Dump of AddressJSONPatch:")
		utils.Dump(body)
	}

	var columns []string
	var args []interface{}
	argIdx := 1

	// Conditionally append fields if they aren't nil
	if body.Street != nil {
		columns = append(columns, fmt.Sprintf("street = $%d", argIdx))
		args = append(args, *body.Street)
		argIdx++
	}

	if body.ExtendedStreet != nil {
		columns = append(columns, fmt.Sprintf("extended_street = $%d", argIdx))
		args = append(args, *body.ExtendedStreet)
		argIdx++
	}

	if body.City != nil {
		columns = append(columns, fmt.Sprintf("city = $%d", argIdx))
		args = append(args, *body.City)
		argIdx++
	}

	if body.State != nil {
		columns = append(columns, fmt.Sprintf("state = $%d", argIdx))
		args = append(args, *body.State)
		argIdx++
	}

	if body.PostalCode != nil {
		columns = append(columns, fmt.Sprintf("postal_code = $%d", argIdx))
		args = append(args, *body.PostalCode)
		argIdx++
	}

	if body.Country != nil {
		columns = append(columns, fmt.Sprintf("country = $%d", argIdx))
		args = append(args, *body.Country)
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

	// If nothing was sent to update, just return the current addresses
	if len(columns) == 0 {
		return d.getAddresses(*body.ContactID) // Helper to get contactID first if needed
	}

	query := fmt.Sprintf(`
        UPDATE addresses 
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
			return nil, fmt.Errorf("address record not found or unauthorized")
		}
		logger.Error("Error patching address: %v", err)
		return nil, fmt.Errorf("failed to patch address: %w", err)
	}

	// Sync token update
	newSyncToken, err := d.IncrementAndGetNewSyncToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to increment sync token: %w", err)
	}

	if err := d.bumpContactSyncToken(contactID, newSyncToken); err != nil {
		logger.Warn("[DATABASE] Failed to bump contact sync token: %v", err)
	}

	return d.getAddresses(contactID)
}

func (d *Database) DeleteContactAddress(userID int, contactID int, addressID int) error {
	logger.Debug("[DATABASE] Begin DeleteContactAddress(userID:%d, contactID:%d, addressID:%d)", userID, contactID, addressID)

	_, err := d.db.Exec("DELETE FROM addresses WHERE id = $1 AND contact_id = $2", addressID, contactID)
	if err != nil {
		logger.Error("[DATABASE] Error deleting Address: %v", err)
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
