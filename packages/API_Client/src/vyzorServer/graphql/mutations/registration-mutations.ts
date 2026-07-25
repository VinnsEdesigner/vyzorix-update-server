// Re-export all mutations from graphql-registration-mutations
export {
  ACK_INBOX,
  DEREGISTER_DEVICE,
} from "../registration/graphql-registration-mutations";

// Additional registration mutations - create placeholder constants for now
// These would need to be implemented based on actual GraphQL schema
import { gql } from "@apollo/client";

export const SUBMIT_REGISTRATION_REQUEST = gql`
  mutation SubmitRegistrationRequest($input: RegistrationRequestInput!) {
    submitRegistrationRequest(input: $input) {
      success
      requestId
    }
  }
`;

export const ACKNOWLEDGE_REQUEST = gql`
  mutation AcknowledgeRequest($requestId: ID!) {
    acknowledgeRequest(requestId: $requestId) {
      success
    }
  }
`;

export const REGISTER_DEVICE = gql`
  mutation RegisterDevice($input: DeviceRegistrationInput!) {
    registerDevice(input: $input) {
      success
      deviceId
    }
  }
`;

export const DISMISS_INBOX_ENTRY = gql`
  mutation DismissInboxEntry($entryId: ID!) {
    dismissInboxEntry(entryId: $entryId) {
      success
    }
  }
`;

export const CONFIRM_REGISTRATION = gql`
  mutation ConfirmRegistration($deviceId: ID!, $confirmationCode: String!) {
    confirmRegistration(deviceId: $deviceId, confirmationCode: $confirmationCode) {
      success
    }
  }
`;
