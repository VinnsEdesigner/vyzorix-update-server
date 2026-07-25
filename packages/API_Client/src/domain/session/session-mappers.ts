import type {
  Session,
  SessionListResponse,
  ConcurrentSessionsResponse,
} from "./session-entity";

export interface RawSession {
  id: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  expires_at: string;
  is_current: boolean;
  selected_organization_id?: string;
}

export interface RawSessionListResponse {
  sessions: RawSession[];
  total: number;
}

export interface RawConcurrentSessionsResponse {
  has_concurrent: boolean;
  count: number;
  sessions: {
    session_id: string;
    ip_address: string;
    user_agent: string;
    created_at: string;
  }[];
}

export function sessionFromRaw(raw: RawSession): Session {
  return {
    id: raw.id,
    ipAddress: raw.ip_address,
    userAgent: raw.user_agent,
    createdAt: new Date(raw.created_at),
    expiresAt: new Date(raw.expires_at),
    isCurrent: raw.is_current,
    selectedOrganizationId: raw.selected_organization_id,
  };
}

export function sessionListFromRaw(raw: RawSessionListResponse): SessionListResponse {
  return {
    sessions: raw.sessions.map(sessionFromRaw),
    total: raw.total,
  };
}

export function concurrentSessionsFromRaw(raw: RawConcurrentSessionsResponse): ConcurrentSessionsResponse {
  return {
    hasConcurrent: raw.has_concurrent,
    count: raw.count,
    sessions: raw.sessions.map((s) => ({
      id: s.session_id,
      ipAddress: s.ip_address,
      userAgent: s.user_agent,
      createdAt: new Date(s.created_at),
      expiresAt: new Date(0),
      isCurrent: false,
    })),
  };
}
