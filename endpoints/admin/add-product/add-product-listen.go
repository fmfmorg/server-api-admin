package addproduct

import "server-api-admin/util/router"

func Listen() {
	router.Router.POST("/admin/add-product", addProduct)
	router.Router.GET("/admin/add-product-init", addProductInit)
}
