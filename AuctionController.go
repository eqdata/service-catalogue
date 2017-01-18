package main

import (
	"net/http"
	"fmt"
	"time"
	"github.com/bradfitz/gomemcache/memcache"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
)

type AuctionController struct { Controller }

// Stores auction data to the Amazon RDS storage once it has been parsed
func (c *AuctionController) store(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Hello :D", r.Body)


	go c.parse() // We don't care when this finishes so run it as an async go process
}

func (c *AuctionController) fetch(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Fetching auction data for item: ", TitleCase(mux.Vars(r)["item_name"], true))

	var auctions Auctions
	existsInCache := true
	encodedItemName := TitleCase(mux.Vars(r)["item_name"], true)

	// Set skip and take parameters
	var skip int
	var take int

	i, err := strconv.Atoi(r.FormValue("skip"))
	if err != nil {
		skip = 0
	} else {
		skip = i
	}
	i, err = strconv.Atoi(r.FormValue("take"))
	if err != nil {
		take = 10
	} else {
		take = i
	}

	fmt.Println("Skip: " + fmt.Sprint(skip) + ", Take: " + fmt.Sprint(take))

	// Attempt to fetch the item from memached
	mc := memcache.New(MC_HOST + ":" + MC_PORT)

	key := "auction:" + encodedItemName + ":s:" + fmt.Sprint(skip) + ":t:" + fmt.Sprint(take)
	mcItem, err := mc.Get(key)
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			fmt.Println("Couldn't find item in the cache")
			existsInCache = false
			auctions = fetchAuctionDataForItem(encodedItemName, skip, take)
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

	// If we still have nothing send back a 404
	fmt.Println("Sending response to client")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(auctions)
}

// Publishes new auction data to Amazon SQS, this service is responsible
// for being the publisher in the pub/sub model, the Relay server
// is the subscriber which streams the data to the consumer via socket.io
func (c *AuctionController) publish() {
	fmt.Println("Pushing data to queue system")
}

//
func (c *AuctionController) parse() {
	// This just emulates that this is now asynchronous
	time.Sleep(2 * time.Second)
	fmt.Println("Parsing the data!")
	c.publish()
}