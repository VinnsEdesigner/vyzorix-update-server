// Contact points domain — generated types + hand-rolled channel constants.
export type {
  ContactPoint,
  ContactPointListResult,
  ContactPointRequest,
  ContactPointTestResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

// ---- Constants (hand-rolled) ----

export const CONTACT_POINT_CHANNELS = ['email', 'webhook', 'slack'] as const;
export type ContactPointChannel = (typeof CONTACT_POINT_CHANNELS)[number];

export function isContactPointChannel(v: string): v is ContactPointChannel {
  return CONTACT_POINT_CHANNELS.includes(v as ContactPointChannel);
}
