package main

import (
	"time"
	"fmt"
	"encoding/json"
)

type Auctions struct {
	Auctions []Auction
}

type Auction struct {
	Seller string
	Item string
	Price float32
	Quantity int32
	Created_at time.Time
	Updated_at time.Time
}

func fetchAuctionDataForItem(itemName string, skip int, take int) Auctions {

	var auctions Auctions

	if take <= 0 { take = 10 }
	if skip <= 0 { skip = 0 }

	query := "SELECT i.name AS itemName, p.name AS sellerName, a.price, a.quantity, a.created_at, a.updated_at " +
		"FROM auctions AS a " +
		"LEFT JOIN players AS p " +
		"ON a.player_id = p.id " +
		"LEFT JOIN items AS i " +
		"ON a.item_id = i.id " +
		"WHERE i.name = ? " +
		"OR i.displayName = ? " +
		"LIMIT ? " +
		"OFFSET ?"


	rows, _ := DB.Query(query, itemName, itemName, take, skip)
	if rows != nil {
		for rows.Next() {
			var a Auction
			err := rows.Scan(&a.Item, &a.Seller, &a.Price, &a.Quantity, &a.Created_at, &a.Updated_at)
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