import { gql } from '@apollo/client';

export const OPERATOR_SETTINGS_FRAGMENT = gql`
  fragment OperatorSettings on OperatorSettings {
    client {
      theme
      language
      timezone
    }
    notifications {
      enabled
      channels
      email {
        thresholdBreach
        deviceOffline
        deviceOnline
        updateAvailable
        commandFailed
      }
      push {
        thresholdBreach
        deviceOffline
        deviceOnline
        updateAvailable
        commandFailed
      }
      webhook {
        enabled
        url
      }
    }
  }
`;

export const DEVICE_SETTINGS_FRAGMENT = gql`
  fragment DeviceSettings on DeviceSettings {
    id
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
    createdAt
    updatedAt
  }
`;

export const ORGANIZATION_SETTINGS_FRAGMENT = gql`
  fragment OrganizationSettings on OrganizationSettings {
    id
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
    createdAt
    updatedAt
  }
`;
