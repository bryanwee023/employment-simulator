package actions

import (
	"encoding/json"
	"log"
)

var ActionGetOrgChart = Action{
	Description: "Get the org chart of the company",
	Parameters:  json.RawMessage(`{}`),
	Execute:     getOrgChart,
}

func getOrgChart(args json.RawMessage) (error, string) {
	log.Println("Fetching org chart")
	// TODO: Implement actual org chart retrieval
	return nil, "{}"
}
