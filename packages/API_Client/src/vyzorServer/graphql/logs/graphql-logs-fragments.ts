export const LOG_ENTRY_FRAGMENT = `
  fragment LogEntry on LogEntry {
    id
    type
    timestamp
    data
  }
`;

export const REGISTRATION_LOG_FRAGMENT = `
  fragment RegistrationLog on RegistrationLog {
    id
    deviceId
    imei
    action
    operatorId
    clientIp
    userAgent
    timestamp
    details
  }
`;

export const DIAGNOSTIC_LOG_FRAGMENT = `
  fragment DiagnosticLog on DiagnosticLog {
    id
    deviceId
    imei
    type
    severity
    message
    timestamp
    metadata
  }
`;
