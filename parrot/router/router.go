package router

import "net/http"

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	return mux
}
