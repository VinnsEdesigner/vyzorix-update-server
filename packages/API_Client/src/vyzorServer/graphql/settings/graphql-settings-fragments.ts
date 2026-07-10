export const SETTINGS_FRAGMENT = /* GraphQL */ `
  fragment Settings on VyzorixSettings {
    operator {
      id
      email
      name
      role
      emailVerified
      createdAt
    }
    connection {
      serverUrl
      deviceId
      dashboardToken
      requestTimeoutMs
      autoReconnect
      strictHmac
    }
    thresholds {
      riskWarn
      riskCrit
      thermalWarn
      thermalCrit
      bufferWarn
      bufferCrit
    }
    notifications {
      enabled
      email {
        enabled
        email
      }
      push {
        enabled
      }
      webhook {
        enabled
        webhookUrl
      }
      events
    }
    advanced {
      logBufferLimit
      signalHistoryLimit
    }
  }
`;

export const OPERATOR_FRAGMENT = /* GraphQL */ `
  fragment Operator on OperatorInfo {
    id
    email
    name
    role
    emailVerified
    createdAt
  }
`;

export const THRESHOLDS_FRAGMENT = /* GraphQL */ `
  fragment Thresholds on ThresholdSettings {
    riskWarn
    riskCrit
    thermalWarn
    thermalCrit
    bufferWarn
    bufferCrit
  }
`;
