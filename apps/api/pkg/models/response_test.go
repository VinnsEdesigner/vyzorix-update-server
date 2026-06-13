package models

import (
	"encoding/json"
	"testing"
)

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{
		Error:   "not_found",
		Message: "The requested resource was not found.",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Error != resp.Error {
		t.Errorf("Error = %q, want %q", unmarshaled.Error, resp.Error)
	}
	if unmarshaled.Message != resp.Message {
		t.Errorf("Message = %q, want %q", unmarshaled.Message, resp.Message)
	}
}

func TestErrorResponse_JSONTags(t *testing.T) {
	data := []byte(`{"error": "bad_request", "message": "Invalid input provided"}`)

	var resp ErrorResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if resp.Error != "bad_request" {
		t.Errorf("Error = %q, want \"bad_request\"", resp.Error)
	}
}

func TestErrorResponse_EmptyMessage(t *testing.T) {
	resp := ErrorResponse{
		Error:   "internal_error",
		Message: "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Message != "" {
		t.Errorf("Message = %q, want \"\"", unmarshaled.Message)
	}
}

func TestOKResponse_JSON(t *testing.T) {
	resp := OKResponse{
		Database:   "ok",
		ServerTime: 1700000000000,
		OK:         true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled OKResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Database != resp.Database {
		t.Errorf("Database = %q, want %q", unmarshaled.Database, resp.Database)
	}
	if unmarshaled.ServerTime != resp.ServerTime {
		t.Errorf("ServerTime = %d, want %d", unmarshaled.ServerTime, resp.ServerTime)
	}
	if unmarshaled.OK != resp.OK {
		t.Errorf("OK = %v, want %v", unmarshaled.OK, resp.OK)
	}
}

func TestOKResponse_Minimal(t *testing.T) {
	resp := OKResponse{
		OK: true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled OKResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.OK != true {
		t.Error("OK should be true")
	}
	if unmarshaled.Database != "" {
		t.Errorf("Database = %q, want \"\"", unmarshaled.Database)
	}
}

func TestOKResponse_AllFields(t *testing.T) {
	resp := OKResponse{
		Database:   "healthy",
		ServerTime: 1700000000000,
		OK:         true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled OKResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !unmarshaled.OK {
		t.Error("OK should be true")
	}
	if unmarshaled.Database != "healthy" {
		t.Errorf("Database = %q, want \"healthy\"", unmarshaled.Database)
	}
}

func TestErrorResponse_AllErrorTypes(t *testing.T) {
	errorTypes := []struct {
		error   string
		message string
	}{
		{"bad_request", "Invalid request format"},
		{"unauthorized", "Authentication required"},
		{"forbidden", "Access denied"},
		{"not_found", "Resource not found"},
		{"conflict", "Resource already exists"},
		{"internal_error", "Something went wrong"},
		{"service_unavailable", "Server is temporarily unavailable"},
	}

	for _, et := range errorTypes {
		resp := ErrorResponse{
			Error:   et.error,
			Message: et.message,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Errorf("json.Marshal(%q) error = %v", et.error, err)
			continue
		}

		var unmarshaled ErrorResponse
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("json.Unmarshal() error = %v", err)
			continue
		}

		if unmarshaled.Error != et.error {
			t.Errorf("Error = %q, want %q", unmarshaled.Error, et.error)
		}
		if unmarshaled.Message != et.message {
			t.Errorf("Message = %q, want %q", unmarshaled.Message, et.message)
		}
	}
}