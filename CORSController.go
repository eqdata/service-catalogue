package main

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
)

type CORSController struct { Controller }

// Stores auction data to the Amazon RDS storage once it has been parsed
func (c *CORSController) reply(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Fetching player: ", mux.Vars(r)["player_name"])
}