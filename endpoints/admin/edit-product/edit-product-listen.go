package editproduct

import "server-api-admin/util/router"

func Listen() {
	router.Router.GET("/admin/edit-product-init/:product-id", editProductInit)
	router.Router.POST("/admin/edit-product", editProduct)
}
