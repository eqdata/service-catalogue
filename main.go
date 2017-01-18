package main

import (
	"net/http"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/gorilla/handlers"
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

	origins     := handlers.AllowedOrigins([]string{"*"})
	credentials := handlers.AllowCredentials()
	methods     := handlers.AllowedMethods([]string{"PUT, OPTIONS, POST, PATCH, DELETE, GET"})
	headers     := handlers.AllowedHeaders([]string{"access-control-allow-origin", "access-control-allow-headers", "x-requested-with"})

	log.Fatal(http.ListenAndServe(":" + PORT, handlers.CORS(origins, credentials, methods, headers)(router)))
}

func cleanup() {
	fmt.Println("Beginning clean-up")

	DB.Close()

	fmt.Println("Finished clean-up")
}