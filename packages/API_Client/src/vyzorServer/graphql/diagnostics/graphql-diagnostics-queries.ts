import { graphqlClient } from '../_shared/graphql-client';
import { DEVICE_INSPECTION_FRAGMENT, TIMELINE_EVENT_FRAGMENT } from './graphql-diagnostics-fragments';

export const GET_DEVICE_INSPECTION = `
  query GetDeviceInspection($imei: String!) {
    deviceInspection(imei: $imei) {
      ...DeviceInspection
    }
  }
  ${DEVICE_INSPECTION_FRAGMENT}
`;

export const GET_DEVICE_TIMELINE = `
  query GetDeviceTimeline($imei: String!, $eventType: TimelineEventType, $startTime: Int, $endTime: Int, $limit: Int, $cursor: String) {
    deviceTimeline(imei: $imei, eventType: $eventType, startTime: $startTime, endTime: $endTime, limit: $limit, cursor: $cursor) {
      events {
        ...TimelineEvent
      }
      hasMore
      nextCursor
    }
  }
  ${TIMELINE_EVENT_FRAGMENT}
`;

export async function queryDeviceInspection(params: { imei: string }) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_INSPECTION,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceTimeline(params: {
  imei: string;
  eventType?: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
  cursor?: string;
}) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_TIMELINE,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
