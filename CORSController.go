package main

import (
	"net/http"
)

type CORSController struct { Controller }

// Stores auction data to the Amazon RDS storage once it has been parsed
func (c *CORSController) reply(w http.ResponseWriter, r  *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With, Access-Control-Request-Headers, Access-Control-Request-Method")

	w.WriteHeader(http.StatusOK)
}