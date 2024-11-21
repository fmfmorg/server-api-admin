package orders

import (
	"context"
	"database/sql"
	"server-api-admin/models"
)

func FetchStatuses(ctx context.Context, tx *sql.Tx) ([]models.Specification, error) {
	orderStatuses := []models.Specification{{
		ID:   0,
		Name: "All statuses",
	}}

	rows, err := tx.QueryContext(
		ctx,
		"SELECT order_status_id, name FROM order_status ORDER BY order_status_id ASC",
	)
	if err != nil && err != sql.ErrNoRows {
		return orderStatuses, err
	}

	defer rows.Close()

	for rows.Next() {
		var s models.Specification
		err = rows.Scan(&s.ID, &s.Name)
		if err != nil {
			return orderStatuses, err
		}
		orderStatuses = append(orderStatuses, s)
	}

	return orderStatuses, nil
}
