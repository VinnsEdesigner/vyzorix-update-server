/**
 * Registration GraphQL Index
 * 
 * Barrel export for all registration GraphQL files.
 */

// Fragments
export {
  INBOX_ENTRY_FRAGMENT,
  DEVICE_FRAGMENT,
  TELEMETRY_FRAME_FRAGMENT,
} from "./graphql-registration-fragments";

// Types
export type {
  RawInboxEntry,
  RawDevice,
  RawTelemetryFrame,
  RawInboxConnection,
  RawDeviceConnection,
  RawTelemetryConnection,
  RegistrationRequestInput,
  RawRegistrationRequestResponse,
  RawAcknowledgeResponse,
  RawRegisterDeviceResponse,
  RawConfirmResponse,
  RawDeregisterResponse,
  RawDismissResponse,
} from "./graphql-registration-types";

// Queries
export {
  GET_INBOX_ENTRIES,
  GET_INBOX_ENTRY,
  GET_DEVICES,
  GET_DEVICE,
  GET_DEVICE_TELEMETRY,
} from "./graphql-registration-queries";

// Mutations
export {
  SUBMIT_REGISTRATION_REQUEST,
  ACKNOWLEDGE_REQUEST,
  REGISTER_DEVICE,
  DISMISS_INBOX_ENTRY,
  CONFIRM_REGISTRATION,
  DEREGISTER_DEVICE,
} from "./graphql-registration-mutations";