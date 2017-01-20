package main

import (
	"strings"
	"fmt"
)

// Generate the display name string from the item given the REST URI Parameter
func TitleCase(name string, urlFriendly bool) string {

	uriParts := strings.Split(name, " ")
	if urlFriendly {
		var newParts []string
		for _, part := range uriParts {
			if strings.Contains(part, "_") {
				subParts := strings.Split(part, "_")
				for _, subPart := range subParts {
					newParts = append(newParts, subPart)
				}
			} else {
				newParts = append(newParts, part)
			}
		}
		uriParts = newParts
	}

	LogInDebugMode("STRING PARTS ARE: ", uriParts)

	uriString := ""
	for _, part := range uriParts {
		compare := strings.ToLower(part)
		if(compare == "the" || compare == "of" || compare == "or" || compare == "and" || compare == "a" || compare == "an" || compare == "on" || compare == "to") {
			part = strings.ToLower(part)
		} else {
			part = strings.Title(strings.ToLower(part))
		}
		if urlFriendly {
			uriString += part + "_"
		} else {
			uriString += part + " "
		}
	}

	uriString = strings.Replace(uriString, "'S", "'s", -1)
	uriString = strings.Replace(uriString, "`S", "`s", -1)
	uriString = uriString[0:len(uriString)-1]
	return uriString
}

// Replaces fmt.Println and is used for logging debug messages
func LogInDebugMode(message string, args ...interface{}) {
	if DEBUG {
		fmt.Println(message, args)
	}
}
