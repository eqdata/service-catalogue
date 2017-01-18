package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func CreateRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		// Decorate the HTTP handler with a server log so we can debug the routes
		var handler http.Handler

		handler = route.handler

		if DEBUG {
			handler = Logger(handler, route.name)
		}

		// Bind the main method
		router.
			Methods(route.method).
			Path(route.pattern).
			Name(route.name).
			Handler(handler)

		// Bind OPTIONS handler for CORS
		var h http.HandlerFunc = CC.reply
		router.
			Methods("OPTIONS").
			Path(route.pattern).
			Name(route.name).
			Handler(h)
	}

	return router
}
