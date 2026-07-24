export interface Session {
  id: string;
  ip_address: string;
  user_agent: string;
  created_at: string;
  expires_at: string;
  is_current: boolean;
  selected_organization_id?: string;
}

export interface SessionListResponse {
  sessions: Session[];
  total: number;
}

export interface ConcurrentSessionsResponse {
  has_concurrent: boolean;
  count: number;
  sessions: Omit<Session, "is_current" | "expires_at">[];
}

export interface RevokeSessionResponse {
  success: boolean;
  message: string;
}

export interface RevokeAllSessionsResponse {
  success: boolean;
  revoked_count: number;
  message: string;
}
