package main

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
)

type ItemController struct { Controller }

func (i *ItemController) fetchItem(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Fetching item: ", TitleCase(mux.Vars(r)["item_name"], true))

	var item Item
	encodedItemName := TitleCase(mux.Vars(r)["item_name"], true)
	item.fetchItemByName(encodedItemName)

	fmt.Println("Item is: ", item)

	json.NewEncoder(w).Encode(item)
}

// Look into caching items in Redis for fast retrieval
func (i *ItemController) fetchItemNamesBySearchString(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fetching item names matching: ", TitleCase(mux.Vars(r)["search_term"], true))

	encodedItemName := TitleCase(mux.Vars(r)["search_term"], true)

	json.NewEncoder(w).Encode(fetchItemsBySubstring(encodedItemName))
}