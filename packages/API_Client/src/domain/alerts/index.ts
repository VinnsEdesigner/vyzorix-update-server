// Re-exports generated alert types + helpers; mappers normalize snake/raw
// payloads the REST layer consumed before codegen took ownership of types.
// Source: packages/API_Client/src/generated/orval-sdk.ts (never edit there).
import type {
	AlertRuleWithInstances,
	AlertHistoryEvent,
	RuleRequest,
	AlertInstance,
} from "../../generated/orval-sdk";
export type {
	AlertRuleWithInstances,
	AlertHistoryEvent,
	RuleRequest,
	AlertInstance,
} from "../../generated/orval-sdk";
export {
	ALERT_CONDITIONS,
	ALERT_METRICS,
	ALERT_STATES,
	type AlertCondition,
	type AlertMetric,
	type AlertState,
} from "../../generated/orval-sdk";

// Domain aliases (canonical naming).
export type AlertRule = AlertRuleWithInstances;
export type AlertEvent = AlertHistoryEvent;
export type { RuleRequest as AlertRuleRequest };

// ---- Raw wire shapes (snake_case from the REST handlers) ----

export interface RawAlertInstance {
	labels: Record<string, string> | null;
	state: string;
	value: number;
	evaluated_at: string | null;
}

export interface RawAlertRule {
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
	on_no_data?: string;
	on_error?: string;
	instances: RawAlertInstance[];
	created_at: string;
	updated_at: string;
}

export interface RawAlertEvent {
	id: string;
	rule_id: string;
	from_state: string;
	to_state: string;
	value: number;
	created_at: number;
}

// ---- Raw → domain mappers (drop unknown enum values to safe defaults) ----

export const alertRuleFromRaw = (raw: RawAlertRule): AlertRuleWithInstances => ({
	id: raw.id,
	org_id: raw.org_id,
	name: raw.name,
	metric: raw.metric,
	condition: raw.condition,
	threshold: raw.threshold,
	for_seconds: raw.for_seconds,
	notify_interval_seconds: raw.notify_interval_seconds,
	enabled: raw.enabled,
	on_no_data: raw.on_no_data ?? "",
	on_error: raw.on_error ?? "",
	webhook_url: raw.webhook_url,
	created_at: raw.created_at,
	updated_at: raw.updated_at,
	instances: raw.instances.map(inst => ({
		labels: inst.labels ?? {},
		state: inst.state,
		value: inst.value,
		evaluated_at: inst.evaluated_at,
	})) as AlertInstance[],
});

export const alertRulesFromRaw = (raw: RawAlertRule[]): AlertRuleWithInstances[] => raw.map(alertRuleFromRaw);

export const alertEventFromRaw = (raw: RawAlertEvent): AlertHistoryEvent => ({
	id: raw.id,
	rule_id: raw.rule_id,
	from_state: raw.from_state,
	to_state: raw.to_state,
	value: raw.value,
	created_at: raw.created_at,
});

export const alertEventsFromRaw = (raw: RawAlertEvent[]): AlertHistoryEvent[] => raw.map(alertEventFromRaw);
