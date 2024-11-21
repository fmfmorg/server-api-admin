package products

import (
	"server-api-admin/util/middlewares"
	"server-api-admin/util/router"
)

func Listen() {
	router.Router.GET("/admin/products", middlewares.Middleware(products))
}
