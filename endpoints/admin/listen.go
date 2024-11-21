package admin

import (
	addproduct "server-api-admin/endpoints/admin/add-product"
	editproduct "server-api-admin/endpoints/admin/edit-product"
	firstemployee "server-api-admin/endpoints/admin/first-employee"
	"server-api-admin/endpoints/admin/order"
	"server-api-admin/endpoints/admin/orders"
	"server-api-admin/endpoints/admin/product"
	"server-api-admin/endpoints/admin/products"
	signin "server-api-admin/endpoints/admin/sign-in"
)

func Listen() {
	addproduct.Listen()
	editproduct.Listen()
	firstemployee.Listen()
	order.Listen()
	orders.Listen()
	product.Listen()
	products.Listen()
	signin.Listen()
}
