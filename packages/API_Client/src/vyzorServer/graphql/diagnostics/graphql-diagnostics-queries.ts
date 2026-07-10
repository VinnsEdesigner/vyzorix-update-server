import { DEVICE_INSPECTION_FRAGMENT, TIMELINE_EVENT_FRAGMENT } from "./graphql-diagnostics-fragments";

export const GET_DEVICE_INSPECTION = /* GraphQL */ `
  query GetDeviceInspection($imei: String!) {
    deviceInspection(imei: $imei) {
      ...DeviceInspection
    }
  }
  ${DEVICE_INSPECTION_FRAGMENT}
`;

export const GET_DEVICE_TIMELINE = /* GraphQL */ `
  query GetDeviceTimeline(
    $imei: String!
    $eventType: TimelineEventType
    $startTime: Int
    $endTime: Int
    $limit: Int
    $cursor: String
  ) {
    deviceTimeline(
      imei: $imei
      eventType: $eventType
      startTime: $startTime
      endTime: $endTime
      limit: $limit
      cursor: $cursor
    ) {
      events {
        ...TimelineEvent
      }
      hasMore
      nextCursor
    }
  }
  ${TIMELINE_EVENT_FRAGMENT}
`;

export const GET_TIMELINE_EVENT = /* GraphQL */ `
  query GetTimelineEvent($id: ID!) {
    timelineEvent(id: $id) {
      ...TimelineEvent
    }
  }
  ${TIMELINE_EVENT_FRAGMENT}
`;