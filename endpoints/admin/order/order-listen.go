package order

import "server-api-admin/util/router"

func Listen() {
	router.Router.POST("/admin/order-init", orderInit)
}
