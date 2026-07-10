/**
 * Registration GraphQL Mutations
 * 
 * GraphQL mutation definitions for device registration.
 * Based on DEVICE_REGISTRATION_SYSTEM.md specification (frontend spec).
 */

// ============================================================================
// Input Types
// ============================================================================

/**
 * RegistrationRequestInput (for device-side submission):
 * {
 *   imei: String!
 *   deviceName: String!
 *   model: String
 *   manufacturer: String
 *   osVersion: String!
 *   appVersion: String!
 *   fcmToken: String!
 *   firmware: String
 *   securityPatch: String
 *   buildId: String
 * }
 */

// ============================================================================
// Mutation Definitions
// ============================================================================

/**
 * Submit registration request (device-side)
 * Mutation: submitRegistrationRequest
 */
export const SUBMIT_REGISTRATION_REQUEST = /* GraphQL */ `
  mutation SubmitRegistrationRequest($input: RegistrationRequestInput!) {
    submitRegistrationRequest(input: $input) {
      success
      status
      messageId
      error
    }
  }
`;

/**
 * Acknowledge request (device-side)
 * Mutation: acknowledgeRequest
 */
export const ACKNOWLEDGE_REQUEST = /* GraphQL */ `
  mutation AcknowledgeRequest($imei: String!) {
    acknowledgeRequest(imei: $imei) {
      success
      status
      error
    }
  }
`;

/**
 * Register device (operator initiated)
 * Mutation: registerDevice
 */
export const REGISTER_DEVICE = /* GraphQL */ `
  mutation RegisterDevice($imei: String!) {
    registerDevice(imei: $imei) {
      success
      status
      deviceId
      message
      error
    }
  }
`;

/**
 * Dismiss inbox entry (operator)
 * Mutation: dismissInboxEntry
 */
export const DISMISS_INBOX_ENTRY = /* GraphQL */ `
  mutation DismissInboxEntry($imei: String!) {
    dismissInboxEntry(imei: $imei) {
      success
      status
      error
    }
  }
`;

/**
 * Confirm registration (device-side)
 * Mutation: confirmRegistration
 */
export const CONFIRM_REGISTRATION = /* GraphQL */ `
  mutation ConfirmRegistration($imei: String!, $confirmed: Boolean!) {
    confirmRegistration(imei: $imei, confirmed: $confirmed) {
      success
      status
      deviceId
      commandSecret
      error
    }
  }
`;

/**
 * Deregister device
 * Mutation: deregisterDevice
 */
export const DEREGISTER_DEVICE = /* GraphQL */ `
  mutation DeregisterDevice($imei: String!) {
    deregisterDevice(imei: $imei) {
      success
      status
      error
    }
  }
`;