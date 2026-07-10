/**
 * Device Fragment
 * 
 * Reusable GraphQL fragment for Device type.
 */

export const DEVICE_FRAGMENT = /* GraphQL */ `
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
