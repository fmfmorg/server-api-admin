package delivery

import (
	"context"
	"database/sql"
	"server-api-admin/models"
)

func FetchDeliveryMethods(ctx context.Context, tx *sql.Tx) ([]models.DeliveryMethod, error) {
	deliveryMethods := make([]models.DeliveryMethod, 0)

	rows, err := tx.QueryContext(
		ctx,
		`
			SELECT dm.delivery_method_id, dm.name, dm.cost, dm.min_spend_for_free, r.name
			FROM delivery_method dm
			LEFT JOIN region r ON dm.region_id = r.region_id
		`,
	)
	if err != nil {
		return deliveryMethods, err
	}

	defer rows.Close()

	for rows.Next() {
		var d models.DeliveryMethod
		var regionName sql.NullString
		err = rows.Scan(
			&d.ID,
			&d.Name,
			&d.Cost,
			&d.MinSpendForFree,
			&regionName,
		)
		if err != nil {
			return deliveryMethods, err
		}
		d.RegionName = regionName.String

		deliveryMethods = append(deliveryMethods, d)
	}

	return deliveryMethods, nil
}
