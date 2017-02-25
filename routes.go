package main

import "net/http"

type Route struct {
	name 	string
	method 	string
	pattern string
	handler http.HandlerFunc
}

type Routes []Route

// Define any application routes here
var routes = Routes {
	Route {
		"Fetch Auctions for Item",
		"GET",
		"/auctions/{item_name}",
		AC.fetch,
	},
	Route {
		"Fetch Item",
		"GET",
		"/items/{item_name}",
		IC.fetchItem,
	},
	Route {
		"Fetch Items by Substring",
		"GET",
		"/items/search/{search_term}",
		IC.fetchItemNamesBySearchString,
	},
	Route {
		"Fetch Player",
		"GET",
		"/players/{player_name}",
		PC.fetch,
	},
}