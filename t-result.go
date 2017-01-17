package main

import (
	"encoding/json"
	"fmt"
)

type Result struct {
	Items []string
};

func (r *Result) serialize() []byte {
	bytes, err := json.Marshal(r)
	if err != nil {
		fmt.Println("ERROR WHEN MARSHALING: ", err)
	}

	LogInDebugMode("Marshalled to: ", bytes)
	return bytes;
}

func (r *Result) deserialize(bytes []byte) Result {
	var res Result

	err := json.Unmarshal(bytes, &res)
	if err != nil {
		fmt.Println("ERROR WHEN UNMARSHALING: ", err)
	}

	LogInDebugMode("Unmarshalled to: ", res)
	return res
}

