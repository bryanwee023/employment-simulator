//go:build e2e

package e2etest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// Tests employees has successfully been created and appears in the org chart.
//
// Requires: RabbitMQ + employee service running (e.g. via docker-compose.test.yml).
func TestEmployees(t *testing.T) {
	officeURL := getOfficeURL()
	resp, err := http.Get(officeURL + "/org-chart")
	if err != nil {
		t.Fatalf("Failed to reach office: %v", err)
	}
	defer resp.Body.Close()

	var employees []Employee
	if err := json.NewDecoder(resp.Body).Decode(&employees); err != nil {
		t.Fatalf("Failed to decode org chart: %v", err)
	}

	if len(employees) != 1 {
		t.Fatalf("Expected 1 employee in org chart, got %d", len(employees))
	}

	for i := range employees {
		employee := employees[i]
		expectedID := fmt.Sprintf("emp-%d", i+1)
		expectedName := fmt.Sprintf("Employee %d", i+1)

		if employee.ID != expectedID {
			t.Errorf("Employee %d: expected ID %q, got %q", i+1, expectedID, employee.ID)
		}
		if employee.Name != expectedName {
			t.Errorf("Employee %d: expected Name %q, got %q", i+1, expectedName, employee.Name)
		}
		if employee.Role != "Staff" {
			t.Errorf("Employee %d: expected Role %q, got %q", i+1, "Staff", employee.Role)
		}

		employeeURL := getEmployeeURL(i)
		meResp, err := http.Get(employeeURL + "/me")
		if err != nil {
			t.Fatalf("Failed to reach employee /me: %v", err)
		}
		defer meResp.Body.Close()

		if meResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200 from /me, got %d", meResp.StatusCode)
		}

		var me Employee
		if err := json.NewDecoder(meResp.Body).Decode(&me); err != nil {
			t.Fatalf("Failed to decode /me response: %v", err)
		}

		if me.ID != employee.ID {
			t.Errorf("/me ID mismatch: expected %q, got %q", employee.ID, me.ID)
		}
		if me.Name != employee.Name {
			t.Errorf("/me Name mismatch: expected %q, got %q", employee.Name, me.Name)
		}
		if me.Role != employee.Role {
			t.Errorf("/me Role mismatch: expected %q, got %q", employee.Role, me.Role)
		}
	}

	t.Logf("Verified all %d employees exist in org chart", len(employees))
}
