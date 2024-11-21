package orders

import (
	"encoding/json"
	"net/http"
	"server-api-admin/util/delivery"
	"server-api-admin/util/orders"
	"server-api-admin/util/postgresdb"

	"github.com/julienschmidt/httprouter"
)

type OrdersInitRequest struct {
	OrderStatus int `json:"orderStatus"`
}

func ordersInit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req OrdersInitRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	orderStatuses, err := orders.FetchStatuses(r.Context(), tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	overviewItems, err := orders.FetchOrderOverview(r.Context(), tx, req.OrderStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	deliveryMethods, err := delivery.FetchDeliveryMethods(r.Context(), tx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"orderStatuses":   orderStatuses,
		"overviewItems":   overviewItems,
		"deliveryMethods": deliveryMethods,
	})
}
