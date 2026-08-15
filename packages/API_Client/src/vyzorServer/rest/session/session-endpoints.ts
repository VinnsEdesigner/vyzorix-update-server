import { restClient } from "../_shared/rest-client";
import {
  sessionListFromRaw,
  concurrentSessionsFromRaw,
  type RawSessionListResponse,
  type RawConcurrentSessionsResponse,
} from "../../../domain/session";
import type {
  SessionListResponse,
  ConcurrentSessionsResponse,
  RevokeSessionResponse,
  RevokeAllSessionsResponse,
} from "../../../domain/session";

const PATHS = {
  sessions: "/v1/auth/sessions",
  sessionsConcurrent: "/v1/auth/sessions/concurrent",
  sessionsRevokeAll: "/v1/auth/sessions/revoke-all",
} as const;

export const sessions = {
  async listSessions(): Promise<SessionListResponse> {
    const response = await restClient.get<RawSessionListResponse>(PATHS.sessions);
    return sessionListFromRaw(response);
  },

  async getConcurrent(): Promise<ConcurrentSessionsResponse> {
    const response = await restClient.get<RawConcurrentSessionsResponse>(PATHS.sessionsConcurrent);
    return concurrentSessionsFromRaw(response);
  },

  async revokeSession(sessionId: string): Promise<RevokeSessionResponse> {
    return restClient.delete<RevokeSessionResponse>(`${PATHS.sessions}/${sessionId}`);
  },

  async revokeAllExceptCurrent(): Promise<{ success: boolean; revoked_count: number; message: string }> {
    return restClient.delete<{ success: boolean; revoked_count: number; message: string }>(PATHS.sessions);
  },

  async revokeAllDevices(): Promise<RevokeAllSessionsResponse> {
    return restClient.post<RevokeAllSessionsResponse>(PATHS.sessionsRevokeAll);
  },
};
