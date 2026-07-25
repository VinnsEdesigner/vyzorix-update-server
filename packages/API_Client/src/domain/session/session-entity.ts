export interface Session {
  id: string;
  ipAddress: string;
  userAgent: string;
  createdAt: Date;
  expiresAt: Date;
  isCurrent: boolean;
  selectedOrganizationId?: string;
}

export interface SessionListResponse {
  sessions: Session[];
  total: number;
}

export interface ConcurrentSessionsResponse {
  hasConcurrent: boolean;
  count: number;
  sessions: Omit<Session, "isCurrent" | "expiresAt">[];
}

export interface RevokeSessionResponse {
  success: boolean;
  message: string;
}

export interface RevokeAllSessionsResponse {
  success: boolean;
  revokedCount: number;
  message: string;
}
