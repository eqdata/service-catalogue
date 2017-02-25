package main

import (
	"net/http"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"strings"
)

type AuctionController struct { Controller }

func (c *AuctionController) fetch(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Fetching auction data for item: ", TitleCase(mux.Vars(r)["item_name"], true))

	fmt.Println("Sending response to client")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "accept, content-type, x-xsrf-token, x-csrf-token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var auctions Auctions
	existsInCache := true
	encodedItemName := TitleCase(mux.Vars(r)["item_name"], true)

	// Set skip and take parameters
	var skip int
	var take int

	server := strings.TrimSpace(strings.ToLower(r.FormValue("server")))
	fmt.Println("Server is: ", server)
	if server != "red" && server != "blue" {
		w.WriteHeader(400)
		w.Write([]byte("The name of the server needs to be specified, please send either red of blue"))
		return
	}

	v, err := strconv.Atoi(r.FormValue("skip"))
	if err != nil {
		skip = 0
	} else {
		skip = v
	}
	v, err = strconv.Atoi(r.FormValue("take"))
	if err != nil {
		take = 10
	} else {
		take = v
	}

	fmt.Println("Skip: " + fmt.Sprint(skip) + ", Take: " + fmt.Sprint(take))

	// Attempt to fetch the item from memached
	mc := memcache.New(MC_HOST + ":" + MC_PORT)

	key := "auction:" + encodedItemName + ":s:" + fmt.Sprint(skip) + ":t:" + fmt.Sprint(take) + ":s:" + server
	mcItem, err := mc.Get(key)
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			fmt.Println("Couldn't find item in the cache")
			existsInCache = false
			auctions = fetchAuctionDataForItem(server, encodedItemName, skip, take)
		} else {
			fmt.Println("Error was: ", err.Error())
			return
		}
	} else if mcItem != nil {
		LogInDebugMode("Got item from memcached: ", mcItem)
		auctions = auctions.deserialize(mcItem.Value)
	}

	// Set the item in memcached regardless of result for 15 minutes
	if !existsInCache {
		fmt.Println("Setting item: " + key + " in cache for: " + fmt.Sprint(AUCTION_CACHE_TIME_IN_SECS) + " seconds")
		mc.Set(&memcache.Item{Key: fmt.Sprint(key), Value: auctions.serialize(), Expiration: AUCTION_CACHE_TIME_IN_SECS})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auctions)
}