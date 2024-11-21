package addproduct

import (
	"encoding/json"
	"net/http"
	"server-api-admin/util/postgresdb"
	"server-api-admin/util/products"

	"github.com/julienschmidt/httprouter"
)

func addProductInit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	materials, metalColors, productTypes, err := products.FetchSpecs(r.Context(), tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"materials":    materials,
		"metalColors":  metalColors,
		"productTypes": productTypes,
	})
}
