package products

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"server-api-admin/config"
	"server-api-admin/util/postgresdb"

	"github.com/julienschmidt/httprouter"
)

type Product struct {
	ProductName     string `json:"name"`
	ProductID       string `json:"id"`
	AdminImageURL   string `json:"adminImageUrl"`
	AdminImageExt   string `json:"adminImageExt"`
	ProductImageURL string `json:"productImageUrl"`
	ProductImageExt string `json:"productImageExt"`
	Price           int    `json:"price"`
	DiscountedPrice int    `json:"discountedPrice"`
}

func products(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	// Fetch products from the database

	rows, err := postgresdb.DB.Query(`
		SELECT
			p.name,
			p.product_id,
			COALESCE(api.file_name, '') AS admin_image_url,
			COALESCE(api.ext, '') AS admin_image_ext,
			COALESCE(pi.file_name, '') AS product_image_url,
			COALESCE(pi.ext, '') AS product_image_ext,
			p.price,
			COALESCE(p.price - d.amount, p.price) AS discounted_price
		FROM
			product p
		LEFT JOIN
			admin_product_image api ON p.product_id = api.product_id AND api.sort_order = (SELECT MIN(sort_order) FROM admin_product_image WHERE product_id = p.product_id)
		LEFT JOIN
			product_image pi ON p.product_id = pi.product_id AND pi.sort_order = (SELECT MIN(sort_order) FROM product_image WHERE product_id = p.product_id)
		LEFT JOIN
			(SELECT
				dd.product_id,
				dd.amount
			FROM
				discount dd
			WHERE
				NOW() BETWEEN dd.start_date AND dd.end_date
			ORDER BY
				dd.start_date DESC
			LIMIT 1) d ON p.product_id = d.product_id
		WHERE
			(p.is_retired = FALSE
			OR EXISTS (SELECT 1 FROM inventory_stock WHERE item_id = p.product_id AND location_id = 1 AND quantity > 0))
		GROUP BY
			p.name, p.product_id, api.file_name, api.ext, pi.file_name, pi.ext, p.price, d.amount
		ORDER BY
			p.product_id
	`)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ProductName,
			&p.ProductID,
			&p.AdminImageURL,
			&p.AdminImageExt,
			&p.ProductImageURL,
			&p.ProductImageExt,
			&p.Price,
			&p.DiscountedPrice,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.AdminImageURL = config.ImageDestProtocol + filepath.Join(config.ImageDestDir, "admin", p.ProductID, p.AdminImageURL)
		p.ProductImageURL = config.ImageDestProtocol + filepath.Join(config.ImageDestDir, "public", p.ProductID, p.ProductImageURL)
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]Product{"products": products})
}
