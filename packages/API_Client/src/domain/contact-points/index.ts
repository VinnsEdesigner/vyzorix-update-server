export const CONTACT_POINT_CHANNELS = ["email", "webhook", "slack"] as const;
export type ContactPointChannel = (typeof CONTACT_POINT_CHANNELS)[number];

export interface ContactPoint {
  id: string;
  orgId: string;
  name: string;
  channel: ContactPointChannel;
  secret: boolean;
  config: Record<string, string>;
  templateId?: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ContactPointRequest {
  name: string;
  channel?: ContactPointChannel;
  secret?: string;
  config?: Record<string, string>;
  templateId?: string;
  enabled?: boolean;
}

interface RawContactPoint {
  id: string;
  org_id: string;
  name: string;
  channel: string;
  secret: boolean;
  config: Record<string, string>;
  template_id?: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

function isChannel(v: string): v is ContactPointChannel {
  return CONTACT_POINT_CHANNELS.includes(v as ContactPointChannel);
}

export const contactPointFromRaw = (raw: RawContactPoint): ContactPoint => ({
  id: raw.id,
  orgId: raw.org_id,
  name: raw.name,
  channel: isChannel(raw.channel) ? raw.channel : "webhook",
  secret: raw.secret,
  config: raw.config ?? {},
  templateId: raw.template_id,
  enabled: raw.enabled,
  createdAt: raw.created_at,
  updatedAt: raw.updated_at,
});

export const contactPointsFromRaw = (raw: RawContactPoint[]): ContactPoint[] => raw.map(contactPointFromRaw);
