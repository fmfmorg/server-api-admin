package order

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"server-api-admin/util/orders"
	"server-api-admin/util/postgresdb"
	"time"

	"github.com/julienschmidt/httprouter"
)

type OrderInitRequest struct {
	OrderID int `json:"orderID"`
}

type Order struct {
	OrderID                        int     `json:"orderID"`
	OrderStatusID                  int     `json:"orderStatusID"`
	OrderDate                      int64   `json:"orderDate"`
	PaymentMethod                  string  `json:"paymentMethod"`
	TotalOrderAmountExDelivery     int     `json:"totalOrderAmountExDelivery"`
	DeliveryCharge                 int     `json:"deliveryCharge"`
	DeliveryMethod                 string  `json:"deliveryMethod"`
	MemberDiscountAmount           int     `json:"memberDiscountAmount"`
	MemberDiscountRate             float64 `json:"memberDiscountRate"`
	StaffDiscount                  int     `json:"staffDiscount"`
	StorewideDiscountVoucherCode   string  `json:"storewideDiscountVoucherCode"`
	StorewideDiscountVoucherRate   float64 `json:"storewideDiscountVoucherRate"`
	StorewideDiscountVoucherAmount int     `json:"storewideDiscountVoucherAmount"`
	Email                          string  `json:"email"`
	FirstName                      string  `json:"firstName"`
	LastName                       string  `json:"lastName"`
	Line1                          string  `json:"lineOne"`
	Line2                          string  `json:"lineTwo"`
	City                           string  `json:"city"`
	StateProvince                  string  `json:"stateProvince"`
	Postcode                       string  `json:"postcode"`
	Country                        string  `json:"country"`
	CollectionPoint                string  `json:"collectionPoint"`
	Region                         string  `json:"region"`
	TrackingNumber                 string  `json:"trackingNumber"`
	DispatchDT                     int64   `json:"dispatchDT"`
	ReceiptDT                      int64   `json:"receiptDT"`
	Details                        string  `json:"orderDetails"`
}

func orderInit(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req OrderInitRequest

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

	var order Order
	var orderDate time.Time
	var dispatchDT, receiptDT sql.NullTime

	err = tx.QueryRowContext(
		r.Context(),
		`
			SELECT 
				co.order_id,
				co.order_status_id,
				co.order_date,
				pm.name as payment_method,
				total_amount_ex_delivery,
				delivery_charge,
				dm.name as delivery_method,
				member_discount_amount,
				COALESCE(mdc.discount_rate, 0) as member_discount_rate,
				staff_discount,
				COALESCE(vspd.coupon_code,'') as storewide_discount_code,
				COALESCE(vspd.discount_rate,0) as storewide_discount_rate,
				COALESCE(storewide_pct_coupon_discount_amount,0),
				email,
				first_name,
				last_name,
				co.line1,
				COALESCE(co.line2,''),
				COALESCE(co.city,''),
				COALESCE(co.state_province,''),
				COALESCE(co.postcode,''),
				rc.country,
				COALESCE(r.name,'') as region,
				COALESCE(la.name,'') as collection_point,
				COALESCE(tracking_number,''),
				dispatch_datetime,
				receipt_datetime,
				COALESCE(details,'')
			FROM customer_order co
			LEFT JOIN payment_method pm ON pm.payment_method_id = co.payment_method_id
			LEFT JOIN delivery_method dm ON dm.delivery_method_id = co.delivery_method_id
			LEFT JOIN vouchers_storewide_percentage_discount vspd ON vspd.id = co.storewide_pct_coupon_id
			LEFT JOIN region_country rc ON rc.region_country_id = co.region_country_id
			LEFT JOIN region r ON r.region_id = rc.region_id
			LEFT JOIN location_address la ON la.address_id = co.collection_point
			LEFT JOIN member_discount_coupon mdc ON mdc.order_id_claimed_with = co.order_id
			WHERE order_id = $1
		`,
		req.OrderID,
	).Scan(
		&order.OrderID,
		&order.OrderStatusID,
		&orderDate,
		&order.PaymentMethod,
		&order.TotalOrderAmountExDelivery,
		&order.DeliveryCharge,
		&order.DeliveryMethod,
		&order.MemberDiscountAmount,
		&order.MemberDiscountRate,
		&order.StaffDiscount,
		&order.StorewideDiscountVoucherCode,
		&order.StorewideDiscountVoucherRate,
		&order.StorewideDiscountVoucherAmount,
		&order.Email,
		&order.FirstName,
		&order.LastName,
		&order.Line1,
		&order.Line2,
		&order.City,
		&order.StateProvince,
		&order.Postcode,
		&order.Country,
		&order.Region,
		&order.CollectionPoint,
		&order.TrackingNumber,
		&dispatchDT,
		&receiptDT,
		&order.Details,
	)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order.OrderDate = orderDate.UnixMilli()

	if dispatchDT.Valid {
		order.DispatchDT = dispatchDT.Time.UnixMilli()
	}

	if receiptDT.Valid {
		order.ReceiptDT = receiptDT.Time.UnixMilli()
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"orderStatuses": orderStatuses,
		"order":         order,
	})
}
