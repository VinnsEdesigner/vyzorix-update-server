// GraphQL Mutations for Vyzorix API
// These replace corresponding REST endpoints

import { gql } from 'graphql-request';

// ============================================================
// DEVICE MUTATIONS
// ============================================================

export const UPDATE_FCM_TOKEN = gql`
  mutation UpdateFCMToken($deviceId: ID!, $token: String!) {
    updateFCMToken(deviceId: $deviceId, token: $token) {
      id
      fcmToken
    }
  }
`;

export const DELETE_DEVICE = gql`
  mutation DeleteDevice($id: ID!) {
    deleteDevice(id: $id)
  }
`;

// ============================================================
// COMMAND MUTATIONS
// ============================================================

export const SEND_COMMAND = gql`
  mutation SendCommand($deviceId: ID!, $command: String!, $args: JSON) {
    sendCommand(deviceId: $deviceId, command: $command, args: $args) {
      dispatchId
      commandId
      deviceId
      command
      status
      createdAt
      deviceOnline
    }
  }
`;

export const RETRY_COMMAND = gql`
  mutation RetryCommand($dispatchId: ID!) {
    retryCommand(dispatchId: $dispatchId) {
      dispatchId
      commandId
      status
      createdAt
    }
  }
`;

export const CANCEL_COMMAND = gql`
  mutation CancelCommand($dispatchId: ID!) {
    cancelCommand(dispatchId: $dispatchId)
  }
`;

// ============================================================
// DEVICE CONTROL MUTATIONS
// ============================================================

export const DISCONNECT_DEVICE = gql`
  mutation DisconnectDevice($deviceId: ID!) {
    disconnectDevice(deviceId: $deviceId)
  }
`;

// ============================================================
// PRE-DEFINED COMMANDS
// These match the REST API commands
// ============================================================

export const COMMAND_MUTATIONS = {
  FORCE_SPEAKER: (deviceId: string) => ({
    deviceId,
    command: 'FORCE_SPEAKER',
    args: null,
  }),
  RESET_AUDIO_HAL: (deviceId: string) => ({
    deviceId,
    command: 'RESET_AUDIO_HAL',
    args: null,
  }),
  TOGGLE_CAPTURE: (deviceId: string) => ({
    deviceId,
    command: 'TOGGLE_CAPTURE',
    args: null,
  }),
  REINIT_PROJECTION: (deviceId: string) => ({
    deviceId,
    command: 'REINIT_PROJECTION',
    args: null,
  }),
  REQUEST_STATUS: (deviceId: string) => ({
    deviceId,
    command: 'REQUEST_STATUS',
    args: null,
  }),
  WAKE_UP_UPDATER: (deviceId: string) => ({
    deviceId,
    command: 'WAKE_UP_UPDATER',
    args: null,
  }),
  DUMP_FLIGHT_DATA: (deviceId: string) => ({
    deviceId,
    command: 'DUMP_FLIGHT_DATA',
    args: null,
  }),
  ROTATE_KEYS: (deviceId: string) => ({
    deviceId,
    command: 'ROTATE_KEYS',
    args: null,
  }),
} as const;
