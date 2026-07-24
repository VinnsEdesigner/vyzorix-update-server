import { LOG_ENTRY_FRAGMENT } from "./graphql-logs-fragments";

export const SUBSCRIBE_TO_LOGS =  `
  subscription SubscribeToLogs($imei: String!, $types: [String!]) {
    logsAdded(imei: $imei, types: $types) {
      ...LogEntry
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;

export const SUBSCRIBE_TO_DEVICE_LOGS =  `
  subscription SubscribeToDeviceLogs($imei: String!) {
    logsAdded(imei: $imei) {
      ...LogEntry
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;
