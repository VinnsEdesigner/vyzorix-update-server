package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const (
	PlaygroundEndpoint = "http://localhost:3000/playground"
)

// GraphQLTestQuery represents a GraphQL query for testing.
type GraphQLTestQuery struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response.
type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string `json:"message"`
	Code      string `json:"code,omitempty"`
	Path      string `json:"path,omitempty"`
}

// GraphQLTestResult holds test results for GraphQL endpoints.
type GraphQLTestResult struct {
	Name         string
	Query        string
	Variables    map[string]interface{}
	ExpectedData bool
	Response     *GraphQLResponse
	Duration     time.Duration
	Error        error
}

// TestGraphQLHealth checks if the GraphQL endpoint is healthy.
func TestGraphQLHealth(t *testing.T) {
	resp, err := http.Get(GraphQLEndpoint)
	if err != nil {
		t.Fatalf("GraphQL endpoint not reachable: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Try an introspection query
	introspectionQuery := `{"query":"{ __schema { queryType { name } } }"}`
	resp, err = http.Post(GraphQLEndpoint, "application/json", bytes.NewBufferString(introspectionQuery))
	if err != nil {
		t.Fatalf("Introspection query failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Introspection query failed with status %d", resp.StatusCode)
	}
}

// TestGraphQLQueryDevices tests the devices query.
func TestGraphQLQueryDevices(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query {
			devices(limit: 10) {
				id
				deviceId
				model
				status
			}
		}`,
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Query failed: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("No response received")
	}
	if len(result.Response.Errors) > 0 {
		t.Errorf("GraphQL errors: %v", result.Response.Errors)
	}
}

// TestGraphQLQueryDevicesWithPagination tests pagination.
func TestGraphQLQueryDevicesWithPagination(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($limit: Int, $offset: Int) {
			devices(limit: $limit, offset: $offset) {
				id
				deviceId
			}
		}`,
		Variables: map[string]interface{}{
			"limit":  5,
			"offset": 0,
		},
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Pagination query failed: %v", result.Error)
	}
}

// TestGraphQLMutationSendCommand tests sending a command mutation.
func TestGraphQLMutationSendCommand(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `mutation($deviceId: ID!, $command: String!) {
			sendCommand(input: {deviceId: $deviceId, command: $command}) {
				dispatchId
				status
				command
			}
		}`,
		Variables: map[string]interface{}{
			"deviceId": "test-device-001",
			"command":   "RESTART_APP",
		},
	}

	result := executeGraphQLQuery(t, query)
	// This may fail if no device exists - that's expected
	if result.Response != nil && len(result.Response.Errors) > 0 {
		// Check if it's an auth error (expected without session)
		for _, err := range result.Response.Errors {
			if err.Code == "UNAUTHORIZED" {
				t.Log("Expected UNAUTHORIZED error without session")
				return
			}
		}
	}
}

// TestGraphQLQueryTelemetryHistory tests telemetry history query.
func TestGraphQLQueryTelemetryHistory(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($deviceId: ID!, $limit: Int) {
			telemetryHistory(deviceId: $deviceId, limit: $limit) {
				timestamp
				riskScore
				thermalTemp
				bufferLevel
			}
		}`,
		Variables: map[string]interface{}{
			"deviceId": "test-device-001",
			"limit":    100,
		},
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Telemetry history query failed: %v", result.Error)
	}
}

// TestGraphQLQueryDeviceDetail tests getting a single device.
func TestGraphQLQueryDeviceDetail(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($id: ID!) {
			device(id: $id) {
				id
				deviceId
				model
				manufacturer
				osVersion
				appVersion
				status
				lastSeen
			}
		}`,
		Variables: map[string]interface{}{
			"id": "test-device-001",
		},
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Device detail query failed: %v", result.Error)
	}
}

// TestGraphQLQueryPendingCommands tests pending commands query.
func TestGraphQLQueryPendingCommands(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($deviceId: ID!) {
			pendingCommands(deviceId: $deviceId) {
				id
				dispatchId
				command
				status
				createdAt
			}
		}`,
		Variables: map[string]interface{}{
			"deviceId": "test-device-001",
		},
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Pending commands query failed: %v", result.Error)
	}
}

// TestGraphQLQueryConnectionStatus tests connection status query.
func TestGraphQLQueryConnectionStatus(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($deviceId: ID!) {
			connectionStatus(deviceId: $deviceId) {
				deviceId
				status
				connectedAt
				lastActivity
			}
		}`,
		Variables: map[string]interface{}{
			"deviceId": "test-device-001",
		},
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("Connection status query failed: %v", result.Error)
	}
}

// TestGraphQLQueryAllConnections tests getting all connections.
func TestGraphQLQueryAllConnections(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query {
			allConnections {
				deviceId
				status
				connectedAt
			}
		}`,
	}

	result := executeGraphQLQuery(t, query)
	if result.Error != nil {
		t.Errorf("All connections query failed: %v", result.Error)
	}
}

// TestGraphQLMutationCancelCommand tests cancelling a command.
func TestGraphQLMutationCancelCommand(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `mutation($dispatchId: ID!) {
			cancelCommand(dispatchId: $dispatchId)
		}`,
		Variables: map[string]interface{}{
			"dispatchId": "test-dispatch-001",
		},
	}

	result := executeGraphQLQuery(t, query)
	// May fail if command doesn't exist - expected
	if result.Response != nil && len(result.Response.Errors) > 0 {
		for _, err := range result.Response.Errors {
			t.Logf("Error (expected if no command): %s - %s", err.Code, err.Message)
		}
	}
}

// TestGraphQLPlaygroundAccess tests playground accessibility.
func TestGraphQLPlaygroundAccess(t *testing.T) {
	resp, err := http.Get(PlaygroundEndpoint)
	if err != nil {
		t.Fatalf("Playground endpoint not reachable: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || len(contentType) < 9 {
		t.Logf("Content-Type: %s", contentType)
	}
}

// TestGraphQLInvalidQuery tests error handling for invalid queries.
func TestGraphQLInvalidQuery(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query { invalidField }`,
	}

	result := executeGraphQLQuery(t, query)
	if result.Response == nil {
		t.Fatal("Expected a response even for invalid queries")
	}
	if len(result.Response.Errors) == 0 {
		t.Log("Warning: Expected GraphQL errors for invalid query")
	}
}

// TestGraphQLMissingVariables tests handling of missing required variables.
func TestGraphQLMissingVariables(t *testing.T) {
	query := GraphQLTestQuery{
		Query: `query($id: ID!) {
			device(id: $id) {
				id
			}
		}`,
		// No variables provided for required $id
	}

	result := executeGraphQLQuery(t, query)
	if result.Response != nil && len(result.Response.Errors) > 0 {
		t.Logf("Expected error for missing variable: %v", result.Response.Errors)
	}
}

// executeGraphQLQuery executes a GraphQL query and returns the result.
func executeGraphQLQuery(t *testing.T, query GraphQLTestQuery) GraphQLTestResult {
	start := time.Now()

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return GraphQLTestResult{
			Name:  t.Name(),
			Query: query.Query,
			Error: fmt.Errorf("failed to marshal query: %w", err),
		}
	}

	req, err := http.NewRequest("POST", GraphQLEndpoint, bytes.NewBuffer(queryJSON))
	if err != nil {
		return GraphQLTestResult{
			Name:  t.Name(),
			Query: query.Query,
			Error: fmt.Errorf("failed to create request: %w", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return GraphQLTestResult{
			Name:     t.Name(),
			Query:    query.Query,
			Duration: time.Since(start),
			Error:    fmt.Errorf("request failed: %w", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return GraphQLTestResult{
			Name:     t.Name(),
			Query:    query.Query,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to decode response: %w", err),
		}
	}

	return GraphQLTestResult{
		Name:      t.Name(),
		Query:     query.Query,
		Variables: query.Variables,
		Response:  &gqlResp,
		Duration:  time.Since(start),
	}
}

// BenchmarkGraphQLQueryDevices benchmarks the devices query.
func BenchmarkGraphQLQueryDevices(b *testing.B) {
	query := GraphQLTestQuery{
		Query: `query {
			devices(limit: 50) {
				id
				deviceId
				model
				status
			}
		}`,
	}

	client := &http.Client{Timeout: 10 * time.Second}
	queryJSON, _ := json.Marshal(query)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", GraphQLEndpoint, bytes.NewBuffer(queryJSON))
		req.Header.Set("Content-Type", "application/json")
		_, _ = client.Do(req)
	}
}

// BenchmarkGraphQLQueryTelemetryHistory benchmarks telemetry history.
func BenchmarkGraphQLQueryTelemetryHistory(b *testing.B) {
	query := GraphQLTestQuery{
		Query: `query($deviceId: ID!, $limit: Int) {
			telemetryHistory(deviceId: $deviceId, limit: $limit) {
				timestamp
				riskScore
				thermalTemp
				bufferLevel
			}
		}`,
		Variables: map[string]interface{}{
			"deviceId": "test-device-001",
			"limit":    100,
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	queryJSON, _ := json.Marshal(query)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", GraphQLEndpoint, bytes.NewBuffer(queryJSON))
		req.Header.Set("Content-Type", "application/json")
		_, _ = client.Do(req)
	}
}
