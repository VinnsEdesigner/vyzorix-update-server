import { gql } from '@apollo/client';

export const TELEMETRY_FRAME_FRAGMENT = gql`
  fragment TelemetryFrame on TelemetryEntry {
    id
    deviceId
    receivedAt
    riskScore
    bufferLevel
    thermalTemp
    payload
  }
`;
