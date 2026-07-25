import { LOG_ENTRY_FRAGMENT } from "./graphql-logs-fragments";

export const SUBSCRIBE_TO_LOGS =  `
  subscription SubscribeToLogs($deviceId: String!, $types: [String!]) {
    logsAdded(deviceId: $deviceId, types: $types) {
      ...LogEntry
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;

export const SUBSCRIBE_TO_DEVICE_LOGS =  `
  subscription SubscribeToDeviceLogs($deviceId: String!) {
    logsAdded(deviceId: $deviceId) {
      ...LogEntry
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;
