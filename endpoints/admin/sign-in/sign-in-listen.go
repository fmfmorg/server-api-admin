package signin

import (
	"server-api-admin/util/middlewares"
	"server-api-admin/util/router"
)

func Listen() {
	router.Router.GET("/admin/sign-in-page-init", middlewares.Middleware(signInPageInit))
	router.Router.POST("/admin/sign-in", middlewares.Middleware(signIn))
}
