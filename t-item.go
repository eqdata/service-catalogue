package main

import (
	"fmt"
	"database/sql"
	"strings"
	"encoding/json"
	"github.com/alexmk92/stringutil"
)

/*
 |------------------------------------------------------------------
 | Type: Item
 |------------------------------------------------------------------
 |
 | Represents an item, when we fetch its data we first attempt to
 | hit our file cache, if the item doesn't exist there we fetch
 | it from the Wiki and then store it to our Mongo store
 |
 | @member name (string): Name of the item (url encoded)
 | @member displayName (string): Name of the item (browser friendly)
 | @member imageSrc (string): URL for the image stored on wiki
 | @member price (float32): The advertised price
 | @member statistics ([]Statistic): An array of all stats for this item
 |
 */

type Item struct {
	Name string
	Image string
	AveragePrice float32
	Statistics []Statistic
	Effect Effect
	Affinities []string
	Races []string
	Classes []string
	Slots []string
}

// Given a search string, find items with a name like this
func fetchItemsBySubstring(searchTerm string) Result {
	//var items []string
	var r Result

	// TODO Put some logic to attempt to fetch from Redis here first...


	// If it doesn't exist in cache, then fetch from DB (the client should enforce its own cache too to prevent spamming server)
	query := "SELECT i.displayName " +
		"FROM items AS i " +
		"WHERE i.name LIKE ? " +
		"LIMIT 15 "

	rows, _ := DB.Query(query, "%" + searchTerm + "%")
	if rows != nil {
		for rows.Next() {
			var name sql.NullString
			err := rows.Scan(&name)
			if err != nil {
				fmt.Println("Scan error for mass item search.")
			}

			if name.Valid && name.String != "" {
				r.Items = append(r.Items, name.String)
			}
		}
		if err := rows.Err(); err != nil {
			fmt.Println("ROW ERROR: ", err.Error())
		}
		DB.CloseRows(rows)
	}

	return r
}

// Given a snake_case string find the item in SQL and populate this struct
func (i *Item) fetchItemByName(itemName string) {
	query := "SELECT i.displayName, i.imageSrc, s.code, s.effect, s.value, e.name as effectName, e.uri, ie.restriction, " +
		"(SELECT AVG(price) FROM auctions WHERE item_id = i.id) AS averagePrice " +
		"FROM items AS i " +
		"LEFT JOIN statistics AS s " +
		"ON s.item_id = i.id " +
		"LEFT JOIN item_effects AS ie " +
		"ON ie.item_id = i.id " +
		"LEFT JOIN effects AS e " +
		"ON ie.effect_id = e.id " +
		"WHERE i.displayName = ? " +
		"OR i.name = ?"

	LogInDebugMode(query)

	rows, _ := DB.Query(query, itemName, itemName)
	if rows != nil {
		for rows.Next() {
			var name, imageSrc, code, effect, effectName, uri, restriction sql.NullString
			var value, averagePrice sql.NullFloat64

			err := rows.Scan(&name, &imageSrc, &code, &effect, &value, &effectName, &uri, &restriction, &averagePrice)
			if err != nil {
				fmt.Println("Scan error: ", err)
			}
			LogInDebugMode("Row is: ", name, imageSrc, code, effect, fmt.Sprint(value), effectName, uri, restriction, fmt.Sprint(averagePrice))

			// If theres an invalid code, trigger a wiki service update?

			// Set the appropriate fields on the struct
			if name.Valid && name.String != "" {
				i.Name = name.String
			}
			if imageSrc.Valid && imageSrc.String != "" {
				i.Image = imageSrc.String
			}
			// Set effect
			if effectName.Valid && effectName.String != "" && i.Effect.Name == "" {
				e := Effect{}
				e.Name = effectName.String
				e.Restriction = restriction.String
				e.URI = WIKI_BASE_URL + uri.String

				i.Effect = e
			}

			 i.setStatistic(code, effect, value)
		}
		if err := rows.Err(); err != nil {
			fmt.Println("ROW ERROR: ", err.Error())
		}
		DB.CloseRows(rows)
	}
}

func (i *Item) setStatistic(code sql.NullString, effect sql.NullString, value sql.NullFloat64) {
	if !code.Valid { return }

	if code.String == "RACE" {
		effect.String = strings.Replace(effect.String, ",", " ", -1)
		effect.String = strings.Replace(effect.String, "  ", " ", -1)
		races := strings.Split(effect.String, " ")

		i.Races = races
	} else if code.String == "CLASS" {
		var classes []string
		effect.String = strings.Replace(effect.String, "  ", " ", -1)
		if stringutil.CaseInsenstiveContains(i.Name, "spell:") {
			classes = strings.Split(effect.String, ",")
		} else {
			effect.String = strings.Replace(effect.String, ",", " ", -1)
			classes = strings.Split(effect.String, " ")
		}

		i.Classes = classes
	} else if code.String == "AFFINITY" {
		parts := stringutil.RegSplit(effect.String, `  +`)
		i.Affinities = parts
	} else if code.String == "SLOT" {
		parts := stringutil.RegSplit(effect.String, `  +`)
		i.Slots = parts
	} else {
		var statistic Statistic
		statistic.Code = strings.ToLower(code.String)

		if effect.Valid == true && effect.String != "" {
			statistic.Value = effect.String
		} else if value.Valid == true && value.Float64 > 0.0 {
			statistic.Value = value.Float64
		}

		i.Statistics = append(i.Statistics, statistic)
	}
}

func (i *Item) serialize() []byte {
	bytes, err := json.Marshal(i)
	if err != nil {
		fmt.Println("ERROR WHEN MARSHALING: ", err)
	}

	LogInDebugMode("Marshalled to: ", bytes)
	return bytes
}

func (i *Item) deserialize(bytes []byte) Item {
	var item Item

	err := json.Unmarshal(bytes, &item)
	if err != nil {
		fmt.Println("ERROR WHEN UNMARSHALING: ", err)
	}

	LogInDebugMode("Unmarshalled to: ", item)
	return item
}