

export const INBOX_ENTRY_FRAGMENT =  `
  fragment InboxEntry on InboxEntry {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    firmware
    securityPatch
    buildId
    fcmToken
    firebaseInstallId
    status
    receivedAt
    updatedAt
    acknowledgedAt
    approvedAt
    rejectedAt
    notes
  }
`;

export const DEVICE_FRAGMENT =  `
  fragment Device on Device {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    fcmToken
    status
    registeredAt
    lastSeen
  }
`;

export const TELEMETRY_FRAME_FRAGMENT =  `
  fragment TelemetryFrame on TelemetryFrame {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
    uptime
  }
`;
