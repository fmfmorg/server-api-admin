package products

import (
	"context"
	"database/sql"
	"server-api-admin/models"
	"strings"
	"time"
)

func FetchProductInfo(ctx context.Context, tx *sql.Tx, productID string) (models.ProductDetails, error) {
	var pd models.ProductDetails
	pd.ID = strings.ToUpper(productID)
	pd.PublicImages = make([]string, 0)
	pd.AdminImages = make([]string, 0)
	pd.Discounts = make([]models.ProductDiscount, 0)

	var createdAt time.Time

	err := tx.QueryRowContext(
		ctx,
		`
			SELECT
				name,
				price,
				material_id,
				metal_color_id,
				description,
				url,
				created_at,
				is_retired,
				product_type_id
			FROM product
			WHERE product_id = $1
		`,
		pd.ID,
	).Scan(
		&pd.Name,
		&pd.Price,
		&pd.MaterialID,
		&pd.MetalColorID,
		&pd.Description,
		&pd.URL,
		&createdAt,
		&pd.IsRetired,
		&pd.ProductTypeID,
	)
	if err != nil {
		return pd, err
	}

	pd.CreatedAt = createdAt.UnixMilli()

	// discounts
	rows, err := tx.QueryContext(
		ctx,
		`
			SELECT
				discount_id,
				amount,
				start_date,
				end_date
			FROM discount
			WHERE product_id = $1
			ORDER BY start_date DESC
		`,
		pd.ID,
	)
	if err != nil {
		return pd, err
	}

	defer rows.Close()

	for rows.Next() {
		var startDT time.Time
		var endDT sql.NullTime
		var d models.ProductDiscount

		err = rows.Scan(
			&d.ID,
			&d.Amount,
			&startDT,
			&endDT,
		)
		if err != nil {
			return pd, err
		}

		d.StartDT = startDT.UnixMilli()
		if endDT.Valid {
			d.EndDT = endDT.Time.UnixMilli()
		}
		pd.Discounts = append(pd.Discounts, d)
	}

	// public images
	rows, err = tx.QueryContext(
		ctx,
		`
			SELECT file_name, ext, catalogue
			FROM product_image
			WHERE product_id = $1
			ORDER BY sort_order ASC
		`,
		pd.ID,
	)
	if err != nil {
		return pd, err
	}

	defer rows.Close()

	for rows.Next() {
		var filename, ext string
		var inCatalogue bool

		err = rows.Scan(&filename, &ext, &inCatalogue)
		if err != nil {
			return pd, err
		}
		if inCatalogue {
			filename = "*" + filename
		}

		pd.PublicImages = append(pd.PublicImages, filename+ext)
	}

	// admin images
	rows, err = tx.QueryContext(
		ctx,
		`
				SELECT file_name, ext
				FROM admin_product_image
				WHERE product_id = $1
				ORDER BY sort_order ASC
			`,
		pd.ID,
	)
	if err != nil {
		return pd, err
	}

	defer rows.Close()

	for rows.Next() {
		var filename, ext string

		err = rows.Scan(&filename, &ext)
		if err != nil {
			return pd, err
		}

		pd.PublicImages = append(pd.PublicImages, filename+ext)
	}

	// stock quantities
	rows, err = tx.QueryContext(
		ctx,
		`
			SELECT
				la.address_id,
				la.name,
				COALESCE(SUM(inv_stock.quantity), 0) AS total_quantity
			FROM
				location_address la
			JOIN
				location l ON la.address_id = l.address_id
			LEFT JOIN
				inventory_stock inv_stock ON l.location_id = inv_stock.location_id AND inv_stock.item_id = $1
			WHERE
				l.type_id = 1
			GROUP BY
				la.address_id, la.name
		`,
		pd.ID,
	)
	if err != nil {
		return pd, err
	}

	defer rows.Close()

	for rows.Next() {
		var s models.StockQuantity
		err = rows.Scan(&s.Address, &s.Name, &s.Quantity)

		if err != nil {
			return pd, err
		}

		pd.StockQuantities = append(pd.StockQuantities, s)
	}

	return pd, nil
}
