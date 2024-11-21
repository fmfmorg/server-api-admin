package editproduct

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"server-api-admin/util/postgresdb"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
)

type Product struct {
	ProductID       string   `json:"productID"`
	Name            string   `json:"name"`
	Price           int      `json:"price"`
	MaterialID      int      `json:"materialID"`
	MetalColorID    int      `json:"metal_colorID"`
	Description     string   `json:"description"`
	URL             string   `json:"url"`
	ProductTypeID   int      `json:"productTypeID"`
	PublicImages    []string `json:"publicImages"`
	AdminImages     []string `json:"adminImages"`
	DiscountStartDT int64    `json:"discountStartDT"`
	DiscountEndDT   int64    `json:"discountEndDT"`
	DiscountAmount  int      `json:"discountAmount"`
	IsRetired       bool     `json:"isRetired"`
}

func editProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	productJSON := r.FormValue("product")
	var product Product
	err := json.Unmarshal([]byte(productJSON), &product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	productID := strings.ToUpper(product.ProductID)

	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	var retiredAt sql.NullTime
	if product.IsRetired {
		retiredAt.Valid = true
		retiredAt.Time = time.Now()
	}

	_, err = tx.ExecContext(
		r.Context(),
		`
			UPDATE product
			SET
				name = $1, 
				price = $2, 
				material_id = $3, 
				metal_color_id = $4, 
				description = $5, 
				url = $6, 
				product_type_id = $7,
				is_retired = $8,
				retired_at = $9
			WHERE
				product_id = $10
		`,
		product.Name,
		product.Price,
		product.MaterialID,
		product.MetalColorID,
		product.Description,
		product.URL,
		product.ProductTypeID,
		product.IsRetired,
		retiredAt,
		productID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(r.Context(), "DELETE FROM product_image WHERE product_id = $1", productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(r.Context(), "DELETE FROM admin_product_image WHERE product_id = $1", productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// public images
	stmt, err := tx.PrepareContext(
		r.Context(),
		pq.CopyIn("product_image", "product_id", "file_name", "ext", "sort_order", "catalogue"),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer stmt.Close()

	for i, filename := range product.PublicImages {
		var inCatalogue bool
		parts := strings.Split(filename, ".")
		ext := fmt.Sprintf(".%s", parts[len(parts)-1])
		filenameNoExt := filename[:len(filename)-len(ext)]

		if len(filenameNoExt) != 0 && filenameNoExt[:1] == "*" {
			filenameNoExt = filenameNoExt[1:]
			inCatalogue = true
		}

		_, err = stmt.ExecContext(
			r.Context(),
			strings.ToUpper(product.ProductID), filenameNoExt, ext, i+1, inCatalogue,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = stmt.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// admin images
	stmt, err = tx.PrepareContext(
		r.Context(),
		pq.CopyIn("admin_product_image", "product_id", "file_name", "ext", "sort_order"),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer stmt.Close()

	for i, filename := range product.PublicImages {
		parts := strings.Split(filename, ".")
		ext := fmt.Sprintf(".%s", parts[len(parts)-1])
		filenameNoExt := filename[:len(filename)-len(ext)]

		_, err = stmt.ExecContext(
			r.Context(),
			strings.ToUpper(product.ProductID), filenameNoExt, ext, i+1,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = stmt.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if product.DiscountAmount != 0 && product.DiscountStartDT != 0 {
		discountStartDT := time.UnixMilli(product.DiscountStartDT)
		var discountEndDT sql.NullTime
		if product.DiscountEndDT != 0 {
			discountEndDT.Valid = true
			discountEndDT.Time = time.UnixMilli(product.DiscountEndDT)
		}

		_, err = tx.ExecContext(
			r.Context(),
			`
				INSERT INTO discount 
				(product_id, amount, start_date, end_date)
				VALUES ($1, $2, $3, $4)
			`,
			productID,
			product.DiscountAmount,
			discountStartDT,
			discountEndDT,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	w.WriteHeader(http.StatusOK)
}
