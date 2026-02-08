package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

func (d *Database) GetLabelID(name string, category string) (int, error) {
	logger.Debug("[DATABASE] Begin GetLabelID(name:%s, category:%s)", name, category)

	query := `
		SELECT id
		FROM contact_label_types
		WHERE name = $1 AND category = $2
	`

	var labelID int

	err := d.db.QueryRow(query, name, category).Scan(&labelID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("unknown name/category combination")
	}
	if err != nil {
		logger.Error("[DATABASE] Error selecting api token: %v", err)
		return 0, fmt.Errorf("failed to validate token: %w", err)
	}

	return labelID, nil
}

func (d *Database) NewLabel(name string, category string) (int, error) {
	logger.Debug("[DATABASE] Begin NewLabel(name:%s, category:%s)", name, category)

	var id int

	// We use DO UPDATE SET name=EXCLUDED.name as a "no-op" trick
	// to ensure RETURNING id always works even on conflict.
	query := `
        INSERT INTO contact_label_types (name, category, is_system)
        VALUES ($1, $2, false)
        ON CONFLICT (name, category) 
        DO UPDATE SET name = EXCLUDED.name 
        RETURNING id`

	err := d.db.QueryRow(query, strings.ToLower(name), strings.ToLower(category)).Scan(&id)
	if err != nil {
		logger.Error("[DATABASE] Error in NewLabel: %v", err)
		return 0, err
	}

	return id, nil
}

func (d *Database) GetLabelMap() (map[int]models.ContactLabelType, error) {
	m, _, _, err := d.getLabelMetadata()
	return m, err
}

func (d *Database) GetLabelReverseMap() (map[string]int, error) {
	_, m, _, err := d.getLabelMetadata()
	return m, err
}

func (d *Database) GetLabelUIMap() (map[string][]models.ContactLabelType, error) {
	_, _, m, err := d.getLabelMetadata()
	return m, err
}

func (d *Database) getLabelMetadata() (
	map[int]models.ContactLabelType, // ID -> Struct (for Export)
	map[string]int, // Key -> ID (for Import)
	map[string][]models.ContactLabelType, // Category -> Slice (for UI)
	error,
) {
	labelMap := make(map[int]models.ContactLabelType)
	revMap := make(map[string]int)
	uiMap := make(map[string][]models.ContactLabelType)

	rows, err := d.db.Query("SELECT id, name, category, is_system FROM contact_label_types ORDER BY is_system DESC, name ASC")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.ContactLabelType
		if err := rows.Scan(&l.ID, &l.Name, &l.Category, &l.IsSystem); err != nil {
			return nil, nil, nil, err
		}

		// 1. For ContactToVCard
		labelMap[l.ID] = l

		// 2. For VCardToContact (Key format: "phone:home")
		key := strings.ToLower(l.Category + ":" + l.Name)
		revMap[key] = l.ID

		// 3. For the UI Templates
		uiMap[l.Category] = append(uiMap[l.Category], l)
	}

	return labelMap, revMap, uiMap, nil
}
