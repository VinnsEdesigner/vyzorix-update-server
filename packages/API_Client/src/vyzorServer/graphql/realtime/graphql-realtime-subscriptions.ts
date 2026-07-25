export const DEVICE_UPDATED_SUBSCRIPTION = `
  subscription OnDeviceUpdated($deviceId: ID) {
    deviceUpdated(deviceId: $deviceId) {
      id
      imei
      deviceName
      status
      lastSeen
    }
  }
`;

export const TELEMETRY_RECEIVED_SUBSCRIPTION = `
  subscription OnTelemetryReceived($deviceId: ID) {
    telemetryReceived(deviceId: $deviceId) {
      id
      deviceId
      receivedAt
      riskScore
      bufferLevel
      thermalTemp
      payload
    }
  }
`;

export const COMMAND_STATUS_SUBSCRIPTION = `
  subscription OnCommandStatusChanged($dispatchId: ID) {
    commandStatusChanged(dispatchId: $dispatchId) {
      dispatchId
      commandId
      deviceId
      command
      status
      createdAt
    }
  }
`;

export const ORGANIZATION_EVENT_SUBSCRIPTION = `
  subscription OnOrganizationEvent($orgId: ID!) {
    organizationEvent(orgId: $orgId) {
      type
      timestamp
      data
    }
  }
`;

export const MEMBER_EVENT_SUBSCRIPTION = `
  subscription OnMemberEvent($orgId: ID!) {
    memberEvent(orgId: $orgId) {
      type
      timestamp
      memberId
      data
    }
  }
`;
