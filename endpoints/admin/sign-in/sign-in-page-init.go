package signin

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func signInPageInit(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}
