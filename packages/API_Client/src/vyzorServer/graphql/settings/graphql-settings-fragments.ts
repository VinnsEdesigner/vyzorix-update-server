export const OPERATOR_SETTINGS_FRAGMENT = `
  fragment OperatorSettings on OperatorSettingsType {
    operator {
      id
      email
      name
      email_verified
      mfa_enabled
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
    device_id
    effective_thresholds {
      temp_min
      temp_max
      battery_min
      battery_max
      speed_max
      distance_max
    }
    organization_thresholds {
      temp_min
      temp_max
      battery_min
      battery_max
      speed_max
      distance_max
    }
  }
`;

export const ORGANIZATION_SETTINGS_FRAGMENT = `
  fragment OrganizationSettings on OrganizationSettingsType {
    organization_id
    default_thresholds {
      temp_min
      temp_max
      battery_min
      battery_max
      speed_max
      distance_max
    }
    updated_at
  }
`;

export const THRESHOLDS_FRAGMENT = `
  fragment Thresholds on ThresholdSettingsType {
    temp_min
    temp_max
    battery_min
    battery_max
    speed_max
    distance_max
  }
`;
