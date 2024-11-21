package addproduct

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server-api-admin/util/postgresdb"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
)

type Product struct {
	ProductID     string   `json:"productID"`
	Name          string   `json:"name"`
	Price         int      `json:"price"`
	MaterialID    int      `json:"materialID"`
	MetalColorID  int      `json:"metal_colorID"`
	Description   string   `json:"description"`
	URL           string   `json:"url"`
	ProductTypeID int      `json:"productTypeID"`
	PublicImages  []string `json:"publicImages"`
	AdminImages   []string `json:"adminImages"`
}

func addProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Extract the product JSON data
	productJSON := r.FormValue("product")
	var product Product
	err := json.Unmarshal([]byte(productJSON), &product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the product data
	if product.ProductID == "" || product.Name == "" || product.Price == 0 || product.MaterialID == 0 || product.MetalColorID == 0 || product.ProductTypeID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	_, err = tx.ExecContext(
		r.Context(),
		`
			INSERT INTO product (product_id, name, price, material_id, metal_color_id, description, url, product_type_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		strings.ToUpper(product.ProductID),
		product.Name,
		product.Price,
		product.MaterialID,
		product.MetalColorID,
		product.Description,
		product.URL,
		product.ProductTypeID,
	)

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

	tx.Commit()

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product and images added successfully"))
}
