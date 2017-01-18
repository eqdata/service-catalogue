package main

import (
	"net/http"
)

type CORSController struct { Controller }

// Stores auction data to the Amazon RDS storage once it has been parsed
func (c *CORSController) reply(w http.ResponseWriter, r  *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	return
}