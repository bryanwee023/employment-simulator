package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const dataFile = "employees.json"

var org *Organization

func main() {
	org = InitOrganization(dataFile)

	http.HandleFunc("/employee/new", newEmployeeHandler)
	http.HandleFunc("/org-chart", orgChartHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("Office service starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ============================
// Handlers
// ============================

func newEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	e := org.NewEmployee()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func orgChartHandler(w http.ResponseWriter, r *http.Request) {
	employees := org.Employees()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok"}`)
}
