package main

import (
	"net/http"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/rs/cors"
)

// Global connection to be used by the server
var DB = Database{}

func main() {
	// Register the cleanup listener:
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	// Initialise DB connections
	fmt.Println("Initialising database connection")
	DB.Open()
	fmt.Println("Connection initialised")

	// Initialise router
	fmt.Println("Starting webserver...")
	fmt.Println("Listening on port: " + PORT)
	router := CreateRouter()

	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":" + PORT, handler))
}

func cleanup() {
	fmt.Println("Beginning clean-up")

	DB.Close()

	fmt.Println("Finished clean-up")
}