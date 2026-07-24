export const IDENTITY_FRAGMENT =  `
  fragment IdentityInfo on IdentityInfo {
    imei
    deviceName
    model
    manufacturer
  }
`;

export const SOFTWARE_FRAGMENT =  `
  fragment SoftwareInfo on SoftwareInfo {
    osVersion
    appVersion
    securityPatch
    buildId
  }
`;

export const REGISTRATION_FRAGMENT =  `
  fragment RegistrationInfo on RegistrationInfo {
    status
    registeredAt
    fcmTokenValid
    fcmTokenRefreshedAt
    commandSecretSet
  }
`;

export const CONNECTION_FRAGMENT =  `
  fragment ConnectionInfo on ConnectionInfo {
    webSocketStatus
    connectedAt
    fcmStatus
    lastSeen
    clientIp
    protocol
  }
`;

export const TELEMETRY_STATS_FRAGMENT =  `
  fragment TelemetryInfo on TelemetryInfo {
    lastTimestamp
    framesToday
    avgLatencyMs
    totalBytesToday
    sessionsToday
  }
`;

export const DEVICE_INSPECTION_FRAGMENT =  `
  fragment DeviceInspection on DeviceInspection {
    identity {
      ...IdentityInfo
    }
    software {
      ...SoftwareInfo
    }
    registration {
      ...RegistrationInfo
    }
    connection {
      ...ConnectionInfo
    }
    telemetry {
      ...TelemetryInfo
    }
  }
  ${IDENTITY_FRAGMENT}
  ${SOFTWARE_FRAGMENT}
  ${REGISTRATION_FRAGMENT}
  ${CONNECTION_FRAGMENT}
  ${TELEMETRY_STATS_FRAGMENT}
`;

export const TIMELINE_EVENT_FRAGMENT =  `
  fragment TimelineEvent on TimelineEvent {
    id
    type
    timestamp
    data
  }
`;
