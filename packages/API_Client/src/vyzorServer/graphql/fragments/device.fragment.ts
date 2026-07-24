

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
