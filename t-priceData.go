package main

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"strings"
	"database/sql"
)

const (
	DAILY = 0
	WEEKLY = 1
	MONTHLY = 2
	YEARLY = 3
	ALL_TIME = 4
)

type PriceData struct {
	Average float32
	Minimum float32
	Maximum float32
	StandardDeviation float32
}

func (p *PriceData) fetchItemPriceStatistics(serverName, itemName string, dRange uint8) {
	fmt.Println("Fetching item: ", TitleCase(itemName, true))

	dateRange := dRange
	if(dateRange > 4) { dRange = 4 }
	if(dateRange < 0) { dRange = 0 }

	existsInCache := true

	// Attempt to fetch the item from memached
	mc := memcache.New(MC_HOST + ":" + MC_PORT)

	var rangeClause string

	switch(dateRange) {
	case DAILY:
		rangeClause = "AND auctions.created_at BETWEEN (CURRENT_DATE() - INTERVAL 1 DAY) AND CURRENT_DATE()"
		break;
	case WEEKLY:
		rangeClause = "AND auctions.created_at BETWEEN (CURRENT_DATE() - INTERVAL 7 DAY) AND CURRENT_DATE()"
		break;
	case MONTHLY:
		rangeClause = "AND auctions.created_at BETWEEN (CURRENT_DATE() - INTERVAL 1 MONTH) AND CURRENT_DATE()"
		break;
	case YEARLY:
		rangeClause = "AND auctions.created_at BETWEEN (CURRENT_DATE() - INTERVAL 1 YEAR) AND CURRENT_DATE()"
		break;
	case ALL_TIME:
		rangeClause = ""
		break;
	}

	fmt.Println("Clause is: ", rangeClause)

	// TODO: Don't set in cache individually, do a group cache afterwards!
	// Use an _ as we don't need to use the cache item returned
	key := "server:" + serverName + ":pricedata:" + itemName + fmt.Sprint(dateRange)
	mcItem, err := mc.Get(key)
	if err != nil {
		if err.Error() == "memcache: cache miss" {
			fmt.Println("Couldn't find item in the cache")
			existsInCache = false

			query := "SELECT " +
				"AVG(NULLIF(price,0)) AS averagePrice, " +
				"STDDEV(NULLIF(price,0)) AS standardDev, " +
				"MIN(NULLIF(price,0)) AS minPrice, " +
				"MAX(NULLIF(price,0)) AS maxPrice " +
				"FROM auctions " +
				"JOIN items AS i " +
				"ON auctions.item_id = i.id " +
				"WHERE (i.displayName = ? " +
				"OR i.name = ?)" +
				rangeClause + " " +
				"AND server ='" + serverName + "' " +
				"LIMIT 1"

			fmt.Println("Query is: " , query)

			rows, _ := DB.Query(query, strings.Replace(itemName, "_", " ", -1), itemName)
			if rows != nil {
				for rows.Next() {
					var averagePrice, standardDev, minPrice, maxPrice sql.NullFloat64

					err := rows.Scan(&averagePrice, &standardDev, &minPrice, &maxPrice)
					if err != nil {
						fmt.Println("Scan error: ", err)
					}
					LogInDebugMode("Row is: ", averagePrice, standardDev, minPrice, maxPrice, fmt.Sprint(averagePrice), fmt.Sprint(standardDev), fmt.Sprint(minPrice), fmt.Sprint(maxPrice))

					// If theres an invalid code, trigger a wiki service update?

					// Set the appropriate fields on the struct
					p.Average = float32(averagePrice.Float64)
					p.StandardDeviation = float32(standardDev.Float64)
					p.Minimum = float32(minPrice.Float64)
					p.Maximum = float32(maxPrice.Float64)

					fmt.Println("PD IS: ", p)
				}
				if err := rows.Err(); err != nil {
					fmt.Println("ROW ERROR: ", err.Error())
				}
				DB.CloseRows(rows)
			}
		} else {
			fmt.Println("Error was: ", err.Error())
			return
		}
	} else if mcItem != nil {
		LogInDebugMode("Got item from memcached: ", mcItem)
		p.deserialize(mcItem.Value)
	}

	// Set the item in memcached regardless of result for 15 minutes
	if !existsInCache {
		fmt.Println("Setting item: " + "item:" + itemName + " in cache for: " + fmt.Sprint((PRICE_CACHE_TIME_IN_SECS)) + " seconds")
		mc.Set(&memcache.Item{Key: fmt.Sprint(key), Value: p.serialize(), Expiration: (PRICE_CACHE_TIME_IN_SECS)})
	} else {

	}
}

func (p *PriceData) serialize() []byte {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println("ERROR WHEN MARSHALING: ", err)
	}

	LogInDebugMode("Marshalled to: ", bytes)
	return bytes
}

func (p *PriceData) deserialize(bytes []byte)  {
	err := json.Unmarshal(bytes, &p)
	if err != nil {
		fmt.Println("ERROR WHEN UNMARSHALING: ", err)
	}

	LogInDebugMode("Unmarshalled to: ", p)
}