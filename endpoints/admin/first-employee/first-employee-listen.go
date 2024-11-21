package firstemployee

import "server-api-admin/util/router"

func Listen() {
	router.Router.POST("/admin/first-employee", createFirstEmployee)
}
