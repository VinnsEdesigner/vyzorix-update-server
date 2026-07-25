export const OPERATOR_SETTINGS_FRAGMENT = `
  fragment OperatorSettings on OperatorSettingsType {
    operator {
      id
      email
      name
      emailVerified
      mfaEnabled
    }
    client {
      theme
      language
      timezone
    }
    notifications {
      enabled
      email {
        enabled
      }
      push {
        enabled
      }
      webhook {
        enabled
        url
      }
    }
  }
`;

export const DEVICE_SETTINGS_FRAGMENT = `
  fragment DeviceSettings on DeviceSettingsType {
    deviceImei
    customName
    location
    thresholds {
      riskWarn
      riskCrit
      thermalWarn
      thermalCrit
      bufferWarn
      bufferCrit
    }
    effectiveThresholds {
      riskWarn
      riskCrit
      thermalWarn
      thermalCrit
      bufferWarn
      bufferCrit
    }
  }
`;

export const ORGANIZATION_SETTINGS_FRAGMENT = `
  fragment OrganizationSettings on OrganizationSettingsType {
    organizationId
    timezone
    dateFormat
    alertCooldownMinutes
    defaultThresholds {
      riskWarn
      riskCrit
      thermalWarn
      thermalCrit
      bufferWarn
      bufferCrit
    }
    updatedAt
  }
`;
