package orders

import (
	"context"
	"database/sql"
	"fmt"
	"server-api-admin/models"
	"time"
)

func FetchOrderOverview(ctx context.Context, tx *sql.Tx, orderStatus int) ([]models.OrderOverviewItem, error) {
	overviewItems := make([]models.OrderOverviewItem, 0)
	subquery := ""
	vars := make([]interface{}, 0)
	if orderStatus != 0 {
		subquery = " AND order_status_id = $1"
		vars = append(vars, orderStatus)
	}
	query := fmt.Sprintf(
		`
			SELECT order_id, order_status_id, order_date, total_amount_ex_delivery, delivery_method_id 
			FROM customer_order
			WHERE 1=1 %s
			ORDER BY order_status_id ASC, order_date ASC 
		`,
		subquery,
	)

	rows, err := tx.QueryContext(ctx, query, vars...)
	if err != nil {
		return overviewItems, err
	}

	defer rows.Close()

	for rows.Next() {
		var s models.OrderOverviewItem
		var orderDate time.Time
		err = rows.Scan(&s.OrderID, &s.OrderStatusID, &orderDate, &s.TotalOrderAmountExDelivery, &s.DeliveryMethodID)
		if err != nil {
			return overviewItems, err
		}

		s.OrderDate = orderDate.UnixMilli()
		overviewItems = append(overviewItems, s)
	}

	return overviewItems, nil
}
