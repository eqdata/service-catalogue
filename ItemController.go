package main

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"strings"
)

type ItemController struct { Controller }

func (i *ItemController) fetchItem(w http.ResponseWriter, r  *http.Request) {
	fmt.Println("Fetching item: ", TitleCase(mux.Vars(r)["item_name"], true))

	server := strings.TrimSpace(strings.ToLower(r.FormValue("server")))
	fmt.Println("Server is: ", server)
	if server != "red" && server != "blue" {
		w.WriteHeader(400)
		w.Write([]byte("The name of the server needs to be specified, please send either red of blue"))
		return
	}

	var item Item
	existsInCache := true

	encodedItemName := TitleCase(strings.Replace(strings.ToLower(mux.Vars(r)["item_name"]), "spell: ", "", -1), true)

	// Attempt to fetch the item from memached
	mc := memcache.New(MC_HOST + ":" + MC_PORT)

	// Use an _ as we don't need to use the cache item returned
	key := "server:" + server + ":item:" + encodedItemName
	mcItem, err := mc.Get(key)
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			fmt.Println("Couldn't find item in the cache")
			existsInCache = false
			item.fetchItemByName(encodedItemName)
		} else {
			fmt.Println("Error was: ", err.Error())
			return
		}
	} else if mcItem != nil {
		LogInDebugMode("Got item from memcached: ", mcItem)
		item = item.deserialize(mcItem.Value)
	} else {
		if item.Name == "" {
			fmt.Println("no item found trying the wiki service")
			res, err := http.Post("http://" + WIKI_SERVICE_HOST + ":" + WIKI_SERVICE_PORT + "/items/" + encodedItemName, "application/json", nil)
			if err == nil && res.StatusCode == 200 {
				item.fetchItemByName(encodedItemName)
			}
		}
	}

	// Set the item in memcached regardless of result for 15 minutes
	if !existsInCache {
		fmt.Println("Setting item: " + fmt.Sprint(key) + " in cache for: " + fmt.Sprint(AUCTION_CACHE_TIME_IN_SECS) + " seconds")
		mc.Set(&memcache.Item{Key: fmt.Sprint(key), Value: item.serialize(), Expiration: AUCTION_CACHE_TIME_IN_SECS})
	}

	item.PriceData = make(map[string]PriceData)

	var daily, weekly, monthly, yearly, allTime PriceData
	weekly.fetchItemPriceStatistics(server, encodedItemName, WEEKLY)
	daily.fetchItemPriceStatistics(server, encodedItemName, DAILY)
	monthly.fetchItemPriceStatistics(server, encodedItemName, MONTHLY)
	yearly.fetchItemPriceStatistics(server, encodedItemName, YEARLY)
	allTime.fetchItemPriceStatistics(server, encodedItemName, ALL_TIME)

	item.PriceData["Weekly"] = weekly
	item.PriceData["Daily"] = daily
	item.PriceData["Monthly"] = monthly
	item.PriceData["Yearly"] = yearly
	item.PriceData["All"] = allTime
	//item.PriceData.fetchItemPriceStatistics(encodedItemName)

	// If we still have nothing send back a 404
	fmt.Println("Sending response to client")
	if item.Name == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type, x-xsrf-token, x-csrf-token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("No item exists with the name: " + encodedItemName + " if you believe this to be an error please contact us."))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type, x-xsrf-token, x-csrf-token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(item)
	}
}

// Look into caching items in Redis for fast retrieval
func (i *ItemController) fetchItemNamesBySearchString(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fetching item names matching: ", TitleCase(mux.Vars(r)["search_term"], true))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "accept, content-type, x-xsrf-token, x-csrf-token")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var results Result

	encodedItemName := TitleCase(mux.Vars(r)["search_term"], true)
	existsInCache := true

	// Attempt to fetch the item from memached
	mc := memcache.New(MC_HOST + ":" + MC_PORT)

	// Use an _ as we don't need to use the cache item returned
	key := "search:" + encodedItemName
	mcItem, err := mc.Get(key)
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			fmt.Println("Couldn't find item in the cache")
			existsInCache = false
			results = fetchItemsBySubstring(encodedItemName)
		} else {
			fmt.Println("Error was: ", err.Error())
			return
		}
	} else if mcItem != nil {
		LogInDebugMode("Got item from memcached: ", mcItem)
		results = results.deserialize(mcItem.Value)
	}

	if !existsInCache {
		fmt.Println("Setting item: " + key + " in cache for: " + fmt.Sprint(SEARCH_CACHE_TIME_IN_SECS) + " seconds")
		mc.Set(&memcache.Item{Key: fmt.Sprint(key), Value: results.serialize(), Expiration: SEARCH_CACHE_TIME_IN_SECS})
	}

	json.NewEncoder(w).Encode(results)
}