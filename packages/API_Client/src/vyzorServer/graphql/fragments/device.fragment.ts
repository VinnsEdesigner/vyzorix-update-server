import { gql } from '@apollo/client';

export const DEVICE_FRAGMENT = gql`
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
