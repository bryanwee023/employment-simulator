package actions

import (
	"encoding/json"
	"log"
)

var ActionDoNothing = Action{
	Description: "Do nothing",
	Parameters: json.RawMessage(`{
		"type": "object",
		"properties": {}
	}`),
	Execute: doNothing,
}

func doNothing(args json.RawMessage) (error, string) {
	log.Println("Doing nothing")
	return nil, ""
}
