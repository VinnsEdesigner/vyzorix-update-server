// Package models contains the public REST and WebSocket payloads shared by the
// Android daemon, dashboard, and Render update server.
//
// Types are organized by domain:
//   - device.go: Device registration, status, and management types
//   - command.go: Command request/response and frame types
//   - telemetry.go: Telemetry data types
//   - updater.go: OTA update manifest types
package models