package orders

import (
	"server-api-admin/util/middlewares"
	"server-api-admin/util/router"
)

func Listen() {
	router.Router.POST("/admin/orders-init", middlewares.Middleware(ordersInit))
}
