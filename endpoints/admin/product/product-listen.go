package product

import (
	"server-api-admin/util/middlewares"
	"server-api-admin/util/router"
)

func Listen() {
	router.Router.GET("/admin/product/:product-id", middlewares.Middleware(productInit))
}
