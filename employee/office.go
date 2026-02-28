package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func fetchEmployee(officeURL string) (Employee, error) {
	resp, err := http.Post(officeURL+"/employee/new", "application/json", nil)
	if err != nil {
		return Employee{}, fmt.Errorf("failed to reach office: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Employee{}, fmt.Errorf("office returned status %d", resp.StatusCode)
	}

	var employee Employee
	if err := json.NewDecoder(resp.Body).Decode(&employee); err != nil {
		return Employee{}, fmt.Errorf("failed to decode employee: %w", err)
	}

	return employee, nil
}
