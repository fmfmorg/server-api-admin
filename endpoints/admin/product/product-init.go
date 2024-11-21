package product

import (
	"encoding/json"
	"net/http"
	"server-api-admin/util/postgresdb"
	"server-api-admin/util/products"

	"github.com/julienschmidt/httprouter"
)

func productInit(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	productID := p.ByName("product-id")

	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	pd, err := products.FetchProductInfo(r.Context(), tx, productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	materials, metalColors, productTypes, err := products.FetchSpecs(r.Context(), tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"product":      pd,
		"materials":    materials,
		"metalColors":  metalColors,
		"productTypes": productTypes,
	})
}
