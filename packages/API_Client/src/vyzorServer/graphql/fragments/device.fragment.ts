export const DEVICE_FRAGMENT = `
  fragment Device on Device {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    status
    registeredAt
    lastSeen
    fcmTokenValid
    commandSecretSet
    connection {
      connected
      lastConnected
      lastDisconnected
    }
  }
`;
