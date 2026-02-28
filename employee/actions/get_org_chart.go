package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var ActionGetOrgChart = Action{
	Description: "Get the org chart of the company",
	Parameters:  json.RawMessage(`{}`),
	Execute:     getOrgChart,
}

func getOrgChart(args json.RawMessage) (error, string) {
	log.Println("Fetching org chart")

	resp, err := http.Get(OfficeURL + "/org-chart")
	if err != nil {
		return fmt.Errorf("failed to reach office: %w", err), ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("office returned status %d", resp.StatusCode), ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err), ""
	}

	return nil, string(body)
}
