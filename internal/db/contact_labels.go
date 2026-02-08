package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
)

func (d *Database) ListLabels() ([]models.ContactLabelType, error) {
	logger.Debug("[DATABASE] Begin ListLabels()")

	var labels []models.ContactLabelType

	// This query aggregates counts from all 4 tables where labels are used
	query := `
        WITH usage_counts AS (
            SELECT label_type_id, COUNT(*) as cnt FROM phones GROUP BY label_type_id
            UNION ALL
            SELECT label_type_id, COUNT(*) as cnt FROM emails GROUP BY label_type_id
            UNION ALL
            SELECT label_type_id, COUNT(*) as cnt FROM addresses GROUP BY label_type_id
            UNION ALL
            SELECT label_type_id, COUNT(*) as cnt FROM urls GROUP BY label_type_id
        ),
        total_counts AS (
            SELECT label_type_id, SUM(cnt) as total FROM usage_counts GROUP BY label_type_id
        )
        SELECT 
            clt.id, 
            clt.name, 
            clt.category, 
            clt.is_system, 
            COALESCE(tc.total, 0) as usage_count
        FROM contact_label_types clt
        LEFT JOIN total_counts tc ON clt.id = tc.label_type_id
        ORDER BY clt.category ASC, clt.is_system DESC, clt.name ASC
    `

	rows, err := d.db.Query(query)
	if err != nil {
		logger.Error("[DATABASE] Error selecting labels with counts: %v", err)
		return nil, fmt.Errorf("error executing query for ListLabels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var l models.ContactLabelType
		err := rows.Scan(&l.ID, &l.Name, &l.Category, &l.IsSystem, &l.UsageCount)
		if err != nil {
			logger.Error("[DATABASE] Error scanning labels: %v", err)
			return nil, fmt.Errorf("error scanning label row: %w", err)
		}
		labels = append(labels, l)
	}

	return labels, nil
}

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

func (d *Database) DeleteContactLabel(labelID int) error {
	var isSystem bool
	var count int
	err := d.db.QueryRow(`
        SELECT is_system, 
        (SELECT COUNT(*) FROM phones WHERE label_type_id = $1) +
        (SELECT COUNT(*) FROM emails WHERE label_type_id = $1) +
		(SELECT COUNT(*) FROM addresses WHERE label_type_id = $1) +
        (SELECT COUNT(*) FROM urls WHERE label_type_id = $1) AS count
        FROM contact_label_types WHERE id = $1`, labelID).Scan(&isSystem, &count)

	if isSystem {
		return errors.New("cannot delete system labels")
	}
	if count > 0 {
		return fmt.Errorf("cannot delete label: used by %d records", count)
	}

	_, err = d.db.Exec("DELETE FROM contact_label_types WHERE id = $1", labelID)
	return err
}
