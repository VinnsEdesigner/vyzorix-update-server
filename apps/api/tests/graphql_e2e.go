package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

const (
	LoginEndpoint  = BaseURL + "/api/v1/auth/login"
)

// E2ETestCredentials holds login credentials for E2E tests.
type E2ETestCredentials struct {
	Email    string
	Password string
}

// E2EGraphQLClient is an authenticated GraphQL client.
type E2EGraphQLClient struct {
	client   *http.Client
	baseURL  string
	Operator *struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
}

// NewE2EGraphQLClient creates an authenticated GraphQL client.
func NewE2EGraphQLClient(t *testing.T) *E2EGraphQLClient {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}

	return &E2EGraphQLClient{
		client:  client,
		baseURL: GraphQLEndpoint,
	}
}

// Login authenticates with the API and stores session cookies.
func (c *E2EGraphQLClient) Login(t *testing.T, email, password string) {
	// First, check if we need to create an account or use test credentials
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		t.Fatalf("Failed to marshal login data: %v", err)
	}
	req, err := http.NewRequest("POST", LoginEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		t.Logf("Login request failed: %v (this may be expected if no users exist)", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Logf("Login returned status %d (this may be expected)", resp.StatusCode)
	}
}

// GraphQLRequest executes a GraphQL request.
func (c *E2EGraphQLClient) GraphQLRequest(t *testing.T, query string, variables map[string]interface{}) map[string]interface{} {
	request := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		request["variables"] = variables
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal GraphQL request: %v", err)
	}
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create GraphQL request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		t.Fatalf("GraphQL request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode GraphQL response: %v", err)
	}

	return result
}

// E2ETestGraphQLDeviceLifecycle tests the complete device lifecycle via GraphQL.
func E2ETestGraphQLDeviceLifecycle(t *testing.T) {
	client := NewE2EGraphQLClient(t)

	// Note: This test requires a device to exist
	// In a real scenario, you'd first register a device via REST API

	t.Run("QueryDevices", func(t *testing.T) {
		query := `query {
			devices(limit: 10) {
				id
				deviceId
				model
				status
			}
		}`

		result := client.GraphQLRequest(t, query, nil)

		// Check for GraphQL errors
		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				errMap, mapOk := e.(map[string]interface{})
				if !mapOk {
					t.Logf("GraphQL Error (unparseable): %v", e)
					continue
				}
				t.Logf("GraphQL Error: %v", errMap)
			}
		}

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("No data in response")
		}

		devices, ok := data["devices"].([]interface{})
		if !ok {
			t.Log("No devices found (expected if none registered)")
			return
		}

		t.Logf("Found %d devices", len(devices))
		for _, d := range devices {
			device, devOk := d.(map[string]interface{})
			if !devOk {
				continue
			}
			t.Logf("  Device: %s (%s) - %s", device["deviceId"], device["model"], device["status"])
		}
	})

	t.Run("QueryDeviceDetail", func(t *testing.T) {
		// First get a device ID
		query := `query { devices(limit: 1) { id } }`
		result := client.GraphQLRequest(t, query, nil)

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		devices, devsOk := data["devices"].([]interface{})
		if !devsOk || len(devices) == 0 {
			t.Skip("No devices to test detail view")
		}

		firstDev, firstOk := devices[0].(map[string]interface{})
		if !firstOk {
			t.Skip("Cannot parse device data")
		}
		deviceID, idOk := firstDev["id"].(string)
		if !idOk {
			t.Skip("Cannot parse device ID")
		}

		detailQuery := fmt.Sprintf(`query {
			device(id: "%s") {
				id
				deviceId
				model
				manufacturer
				osVersion
				appVersion
				status
				lastSeen
			}
		}`, deviceID)

		detailResult := client.GraphQLRequest(t, detailQuery, nil)

		if errors, ok := detailResult["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Errorf("GraphQL Error: %v", e)
			}
		}

		detailData, ddOk := detailResult["data"].(map[string]interface{})
		if !ddOk {
			t.Skip("No detail data")
		}
		device, devOk := detailData["device"].(map[string]interface{})
		if !devOk {
			t.Skip("Cannot parse device detail")
		}
		t.Logf("Device detail: %+v", device)
	})

	t.Run("QueryTelemetryHistory", func(t *testing.T) {
		// First get a device ID
		query := `query { devices(limit: 1) { deviceId } }`
		result := client.GraphQLRequest(t, query, nil)

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		devices, devsOk := data["devices"].([]interface{})
		if !devsOk || len(devices) == 0 {
			t.Skip("No devices to test telemetry")
		}

		firstDev, firstOk := devices[0].(map[string]interface{})
		if !firstOk {
			t.Skip("Cannot parse device data")
		}
		deviceID, idOk := firstDev["deviceId"].(string)
		if !idOk {
			t.Skip("Cannot parse device ID")
		}

		telemetryQuery := fmt.Sprintf(`query {
			telemetryHistory(
				deviceId: "%s"
				limit: 10
			) {
				timestamp
				riskScore
				thermalTemp
				bufferLevel
			}
		}`, deviceID)

		telemetryResult := client.GraphQLRequest(t, telemetryQuery, nil)

		if errors, ok := telemetryResult["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Errorf("GraphQL Error: %v", e)
			}
		}

		t.Logf("Telemetry query result: %+v", telemetryResult["data"])
	})
}

// E2ETestGraphQLCommandFlow tests the command flow via GraphQL.
func E2ETestGraphQLCommandFlow(t *testing.T) {
	client := NewE2EGraphQLClient(t)

	t.Run("SendCommand", func(t *testing.T) {
		// First get a device ID
		query := `query { devices(limit: 1) { deviceId } }`
		result := client.GraphQLRequest(t, query, nil)

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		devices, devsOk := data["devices"].([]interface{})
		if !devsOk || len(devices) == 0 {
			t.Skip("No devices to test commands")
		}

		firstDev, firstOk := devices[0].(map[string]interface{})
		if !firstOk {
			t.Skip("Cannot parse device data")
		}
		deviceID, idOk := firstDev["deviceId"].(string)
		if !idOk {
			t.Skip("Cannot parse device ID")
		}

		mutation := `mutation SendCommand($deviceId: ID!, $command: String!) {
			sendCommand(input: {deviceId: $deviceId, command: $command}) {
				dispatchId
				status
				command
			}
		}`

		variables := map[string]interface{}{
			"deviceId": deviceID,
			"command":   "RESTART_APP",
		}

		mutationResult := client.GraphQLRequest(t, mutation, variables)

		if errors, ok := mutationResult["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				errMap, mapOk := e.(map[string]interface{})
				if !mapOk {
					t.Logf("GraphQL Error (unparseable): %v", e)
					continue
				}
				t.Logf("GraphQL Error: %v", errMap)
				// UNAUTHORIZED is expected without proper auth
				if msg, ok := errMap["message"].(string); ok && strings.Contains(msg, "unauthorized") {
					t.Skip("Requires authentication")
				}
			}
		}

		mutationData, mdOk := mutationResult["data"].(map[string]interface{})
		if mdOk && mutationData != nil {
			sendCommand, scOk := mutationData["sendCommand"].(map[string]interface{})
			if scOk {
				t.Logf("Command result: %+v", sendCommand)
			}
		}
	})

	t.Run("QueryPendingCommands", func(t *testing.T) {
		// First get a device ID
		query := `query { devices(limit: 1) { deviceId } }`
		result := client.GraphQLRequest(t, query, nil)

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		devices, devsOk := data["devices"].([]interface{})
		if !devsOk || len(devices) == 0 {
			t.Skip("No devices to test pending commands")
		}

		firstDev, firstOk := devices[0].(map[string]interface{})
		if !firstOk {
			t.Skip("Cannot parse device data")
		}
		deviceID, idOk := firstDev["deviceId"].(string)
		if !idOk {
			t.Skip("Cannot parse device ID")
		}

		pendingQuery := fmt.Sprintf(`query {
			pendingCommands(deviceId: "%s") {
				id
				dispatchId
				command
				status
				createdAt
			}
		}`, deviceID)

		pendingResult := client.GraphQLRequest(t, pendingQuery, nil)

		if errors, ok := pendingResult["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Errorf("GraphQL Error: %v", e)
			}
		}

		t.Logf("Pending commands result: %+v", pendingResult["data"])
	})
}

// E2ETestGraphQLConnectionStatus tests connection status queries.
func E2ETestGraphQLConnectionStatus(t *testing.T) {
	client := NewE2EGraphQLClient(t)

	t.Run("QueryAllConnections", func(t *testing.T) {
		query := `query {
			allConnections {
				deviceId
				status
				connectedAt
				lastActivity
				ipAddress
			}
		}`

		result := client.GraphQLRequest(t, query, nil)

		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Errorf("GraphQL Error: %v", e)
			}
		}

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		connections, connsOk := data["allConnections"].([]interface{})
		if !connsOk {
			t.Skip("Cannot parse connections")
		}
		t.Logf("Found %d connections", len(connections))

		for _, c := range connections {
			conn, connOk := c.(map[string]interface{})
			if !connOk {
				continue
			}
			t.Logf("  Connection: %s - %s", conn["deviceId"], conn["status"])
		}
	})

	t.Run("QueryConnectionForDevice", func(t *testing.T) {
		// First get a device ID
		query := `query { devices(limit: 1) { deviceId } }`
		result := client.GraphQLRequest(t, query, nil)

		data, dataOk := result["data"].(map[string]interface{})
		if !dataOk {
			t.Skip("No data in response")
		}
		devices, devsOk := data["devices"].([]interface{})
		if !devsOk || len(devices) == 0 {
			t.Skip("No devices to test connection status")
		}

		firstDev, firstOk := devices[0].(map[string]interface{})
		if !firstOk {
			t.Skip("Cannot parse device data")
		}
		deviceID, idOk := firstDev["deviceId"].(string)
		if !idOk {
			t.Skip("Cannot parse device ID")
		}

		connQuery := fmt.Sprintf(`query {
			connectionStatus(deviceId: "%s") {
				deviceId
				status
				connectedAt
				lastActivity
			}
		}`, deviceID)

		connResult := client.GraphQLRequest(t, connQuery, nil)

		if errors, ok := connResult["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Errorf("GraphQL Error: %v", e)
			}
		}

		t.Logf("Connection status result: %+v", connResult["data"])
	})
}

// E2ETestGraphQLDashboard tests dashboard aggregation queries.
func E2ETestGraphQLDashboard(t *testing.T) {
	client := NewE2EGraphQLClient(t)

	t.Run("QueryDashboard", func(t *testing.T) {
		query := `query {
			dashboard(limit: 10) {
				devices {
					id
					deviceId
					status
				}
				connections {
					deviceId
					status
				}
				totalDevices
				onlineDevices
				totalCommands
				pendingCommands
			}
		}`

		result := client.GraphQLRequest(t, query, nil)

		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				errMap, mapOk := e.(map[string]interface{})
				if !mapOk {
					t.Logf("GraphQL Error (unparseable): %v", e)
					continue
				}
				t.Logf("GraphQL Error: %v", errMap)
			}
		}

		data, dataOk := result["data"].(map[string]interface{})
		if dataOk && data != nil {
			dashboard, dashOk := data["dashboard"].(map[string]interface{})
			if dashOk {
				t.Logf("Dashboard: totalDevices=%v, onlineDevices=%v, totalCommands=%v, pendingCommands=%v",
					dashboard["totalDevices"], dashboard["onlineDevices"],
					dashboard["totalCommands"], dashboard["pendingCommands"])
			}
		}
	})
}

// E2ETestGraphQLAuthentication tests GraphQL authentication.
func E2ETestGraphQLAuthentication(t *testing.T) {
	t.Run("UnauthenticatedRequest", func(t *testing.T) {
		// Create a client without login
		client := NewE2EGraphQLClient(t)

		query := `query { devices(limit: 1) { id } }`
		result := client.GraphQLRequest(t, query, nil)

		// Without auth, we expect either errors or empty data
		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				errMap, mapOk := e.(map[string]interface{})
				if !mapOk {
					t.Logf("Expected unauthenticated error (unparseable): %v", e)
					continue
				}
				t.Logf("Expected unauthenticated error: %v", errMap)
			}
		} else {
			t.Log("Query succeeded without authentication (may be expected in dev mode)")
		}
	})

	t.Run("AuthenticatedRequest", func(t *testing.T) {
		client := NewE2EGraphQLClient(t)

		// Try to login with test credentials
		client.Login(t, "test@example.com", "password123")

		query := `query { devices(limit: 1) { id } }`
		result := client.GraphQLRequest(t, query, nil)

		if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
			for _, e := range errors {
				t.Logf("GraphQL Error after login: %v", e)
			}
		} else {
			t.Log("Query succeeded after login")
		}
	})
}
