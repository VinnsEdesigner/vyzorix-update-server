export const ALERT_METRICS = ["device_offline_count", "device_offline_percent", "command_failure_rate"] as const;
export type AlertMetric = (typeof ALERT_METRICS)[number];

export const ALERT_CONDITIONS = ["gt", "gte", "lt", "lte"] as const;
export type AlertCondition = (typeof ALERT_CONDITIONS)[number];

export const ALERT_STATES = ["inactive", "pending", "firing"] as const;
export type AlertState = (typeof ALERT_STATES)[number];

export interface AlertRule {
  id: string;
  orgId: string;
  name: string;
  metric: AlertMetric;
  condition: AlertCondition;
  threshold: number;
  forSeconds: number;
  notifyIntervalSeconds: number;
  webhookUrl: string;
  enabled: boolean;
  state: AlertState;
  value: number;
  evaluatedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface AlertRuleRequest {
  name: string;
  metric: AlertMetric;
  condition: AlertCondition;
  threshold: number;
  forSeconds?: number;
  notifyIntervalSeconds?: number;
  webhookUrl?: string;
  enabled?: boolean;
}

export interface AlertEvent {
  id: string;
  ruleId: string;
  fromState: AlertState;
  toState: AlertState;
  value: number;
  createdAt: number;
}

interface RawAlertRule {
  id: string;
  org_id: string;
  name: string;
  metric: string;
  condition: string;
  threshold: number;
  for_seconds: number;
  notify_interval_seconds: number;
  webhook_url: string;
  enabled: boolean;
  state: string;
  value: number;
  evaluated_at: string | null;
  created_at: string;
  updated_at: string;
}

interface RawAlertEvent {
  id: string;
  rule_id: string;
  from_state: string;
  to_state: string;
  value: number;
  created_at: number;
}

function isAlertMetric(v: string): v is AlertMetric {
  return ALERT_METRICS.includes(v as AlertMetric);
}

function isAlertCondition(v: string): v is AlertCondition {
  return ALERT_CONDITIONS.includes(v as AlertCondition);
}

function isAlertState(v: string): v is AlertState {
  return ALERT_STATES.includes(v as AlertState);
}

export const alertRuleFromRaw = (raw: RawAlertRule): AlertRule => ({
  id: raw.id,
  orgId: raw.org_id,
  name: raw.name,
  metric: isAlertMetric(raw.metric) ? raw.metric : "device_offline_count",
  condition: isAlertCondition(raw.condition) ? raw.condition : "gt",
  threshold: raw.threshold,
  forSeconds: raw.for_seconds,
  notifyIntervalSeconds: raw.notify_interval_seconds,
  webhookUrl: raw.webhook_url,
  enabled: raw.enabled,
  state: isAlertState(raw.state) ? raw.state : "inactive",
  value: raw.value,
  evaluatedAt: raw.evaluated_at,
  createdAt: raw.created_at,
  updatedAt: raw.updated_at,
});

export const alertRulesFromRaw = (raw: RawAlertRule[]): AlertRule[] => raw.map(alertRuleFromRaw);

export const alertEventFromRaw = (raw: RawAlertEvent): AlertEvent => ({
  id: raw.id,
  ruleId: raw.rule_id,
  fromState: isAlertState(raw.from_state) ? raw.from_state : "inactive",
  toState: isAlertState(raw.to_state) ? raw.to_state : "inactive",
  value: raw.value,
  createdAt: raw.created_at,
});

export const alertEventsFromRaw = (raw: RawAlertEvent[]): AlertEvent[] => raw.map(alertEventFromRaw);
