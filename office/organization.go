package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type Organization struct {
	mu        sync.RWMutex
	filePath  string
	employees []Employee
}

type Employee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func InitOrganization(filePath string) *Organization {
	org := &Organization{
		filePath:  filePath,
		employees: []Employee{},
	}
	org.loadEmployees()
	return org
}

// ============================
// Organization Management
// ============================

func (org *Organization) NewEmployee() Employee {
	org.mu.Lock()
	defer org.mu.Unlock()

	e := Employee{
		ID:   fmt.Sprintf("emp-%d", len(org.employees)+1),
		Name: fmt.Sprintf("Employee %d", len(org.employees)+1),
		Role: "Staff",
	}

	log.Printf("Created employee: %s (%s)", e.Name, e.Role)
	org.employees = append(org.employees, e)

	org.save()
	return e
}

func (org *Organization) Employees() []Employee {
	org.mu.RLock()
	defer org.mu.RUnlock()
	return org.employees
}

// ============================
// Persistence
// ============================

func (org *Organization) loadEmployees() {
	data, err := os.ReadFile(org.filePath)
	if err != nil {
		log.Printf("No existing %s, starting with empty list", org.filePath)
		org.employees = []Employee{}
		return
	}

	if err := json.Unmarshal(data, &org.employees); err != nil {
		log.Fatalf("Failed to parse %s: %v", org.filePath, err)
	}
}

// save persists employees to disk. Must be called with mu held.
func (org *Organization) save() error {
	data, err := json.MarshalIndent(org.employees, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal employees: %w", err)
	}
	return os.WriteFile(org.filePath, data, 0644)
}
