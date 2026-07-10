export type InboxStatus = "pending" | "approved" | "rejected";

export type AcknowledgeAction = "approve" | "reject";

export interface InboxEntry {
  id: string;
  imei: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  fcmToken: string;
  firebaseInstallId: string;
  status: InboxStatus;
  createdAt: Date;
  approvedAt?: Date;
  rejectedAt?: Date;
  commandSecret?: string;
  notes?: string;
  operatorId?: string;
}

export interface InboxListItem {
  id: string;
  imei: string;
  model?: string;
  manufacturer?: string;
  status: InboxStatus;
  createdAt: Date;
}

export interface InboxListResult {
  requests: InboxEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface AckResult {
  id: string;
  imei: string;
  status: InboxStatus;
  approvedAt?: Date;
  rejectedAt?: Date;
  commandSecret?: string;
  fcmPushSent?: boolean;
  notes?: string;
}

export interface DeregisterResult {
  imei: string;
  status: string;
  deregisteredAt: Date;
  retentionUntil: Date;
}

export function isInboxTerminal(entry: InboxEntry | InboxListItem): boolean {
  return entry.status === "rejected";
}

export function getInboxStatusLabel(status: InboxStatus): string {
  const labels: Record<InboxStatus, string> = {
    pending: "Pending",
    approved: "Approved",
    rejected: "Rejected",
  };
  return labels[status];
}