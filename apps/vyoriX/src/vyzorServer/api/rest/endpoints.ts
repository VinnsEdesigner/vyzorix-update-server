/**
 * REST Endpoints
 * 
 * Barrel export for all REST API endpoints.
 * Re-exports from feature-specific endpoint files.
 */

// Registration endpoints
export {
  REGISTRATION_PATHS,
  RawInboxEntry,
  RawDevice,
  RawTelemetryFrame,
  InboxStatus,
  DeviceStatus,
  AcknowledgeAction,
  InboxEntry,
  Device,
  TelemetryFrame,
  AcknowledgeRequest,
  AcknowledgeResponse,
  DeregisterResponse,
  RegisterDeviceRequest,
  RegisterDeviceResponse,
  ConfirmRegistrationRequest,
  ConfirmRegistrationResponse,
  TelemetryParams,
  TelemetryResponse,
  fetchInboxEntries,
  fetchInboxEntry,
  acknowledgeInbox,
  dismissInboxEntry,
  fetchDevices,
  fetchDevice,
  deregisterDevice,
  registerDevice,
  confirmRegistration,
  fetchDeviceTelemetry,
} from "./registration-rest";
