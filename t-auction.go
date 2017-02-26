package main

import (
	"time"
	"fmt"
	"encoding/json"
	"strings"
)

type Auctions struct {
	Auctions []Auction
}

type Auction struct {
	Seller string
	Item string
	Price float32
	Quantity int32
	Server string
	Auctioned_At time.Time
	Auction_Line string
}

func fetchAuctionDataForItem(serverName string, itemName string, skip int, take int) Auctions {

	var auctions Auctions

	if take <= 0 { take = 10 }
	if skip <= 0 { skip = 0 }

	fmt.Println("Checking if server is: ", serverName)

	query := "SELECT i.displayName AS itemName, p.name AS sellerName, a.price, a.quantity, a.server, a.created_at, a.raw_auction " +
		"FROM auctions AS a " +
		"LEFT JOIN players AS p " +
		"ON a.player_id = p.id " +
		"LEFT JOIN items AS i " +
		"ON a.item_id = i.id " +
		"WHERE (i.name = ? " +
		"OR i.displayName = ?) " +
		"AND a.server = ? " +
		"ORDER BY a.created_at DESC " +
		"LIMIT ? " +
		"OFFSET ?"

	itemName = strings.Replace(itemName, "_", " ", -1)
	rows, _ := DB.Query(query, itemName, itemName, serverName, take, skip)
	if rows != nil {
		for rows.Next() {
			var a Auction
			err := rows.Scan(&a.Item, &a.Seller, &a.Price, &a.Quantity, &a.Server, &a.Auctioned_At, &a.Auction_Line)
			if err != nil {
				fmt.Println("Scan error: ", err)
			}
			fmt.Println("Appending auction: ", a)
			auctions.Auctions = append(auctions.Auctions, a)
		}
		if err := rows.Err(); err != nil {
			fmt.Println("ROW ERROR: ", err.Error())
		}
		DB.CloseRows(rows)
	}

	return auctions
}

func (a *Auctions) serialize() []byte {
	bytes, err := json.Marshal(a)
	if err != nil {
		fmt.Println("Error when marhsalling auctions: ", err)
	}

	LogInDebugMode("Marshalled to: ", bytes)
	return bytes
}

func (a *Auctions) deserialize(bytes []byte) Auctions {
	var auctions Auctions

	err := json.Unmarshal(bytes, &auctions)
	if err != nil {
		fmt.Println("Error unmarshalling auctions: ", err)
	}

	LogInDebugMode("Unmarshalled to: ", auctions)
	return auctions
}