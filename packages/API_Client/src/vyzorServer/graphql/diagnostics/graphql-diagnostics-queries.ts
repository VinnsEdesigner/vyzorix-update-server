import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';
import { DEVICE_INSPECTION_FRAGMENT, TIMELINE_EVENT_FRAGMENT } from './graphql-diagnostics-fragments';

export const GET_DEVICE_INSPECTION = gql`
  query GetDeviceInspection($imei: String!, $organizationId: ID!) {
    deviceInspection(imei: $imei, organizationId: $organizationId) {
      ...DeviceInspection
    }
  }
  ${DEVICE_INSPECTION_FRAGMENT}
`;

export const GET_DEVICE_TIMELINE = gql`
  query GetDeviceTimeline($imei: String!, $organizationId: ID!, $eventType: TimelineEventType, $startTime: Int, $endTime: Int, $limit: Int, $cursor: String) {
    deviceTimeline(imei: $imei, organizationId: $organizationId, eventType: $eventType, startTime: $startTime, endTime: $endTime, limit: $limit, cursor: $cursor) {
      events {
        ...TimelineEvent
      }
      hasMore
      nextCursor
    }
  }
  ${TIMELINE_EVENT_FRAGMENT}
`;

export async function queryDeviceInspection(params: { imei: string; organizationId: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_INSPECTION,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceTimeline(params: {
  imei: string;
  organizationId: string;
  eventType?: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
  cursor?: string;
}): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_TIMELINE,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
