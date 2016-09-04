package main

import (
	"fmt"
	"testing"

	"encoding/json"
)

func Test_json(t *testing.T) {
	a := TestObj{Id: 1, Name: "test"}
	testJson, _ := json.Marshal(a)
	fmt.Println(string(testJson))
}
