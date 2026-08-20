import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  alertEventFromRaw,
  alertRuleFromRaw,
  alertRulesFromRaw,
  type AlertRule,
  type AlertEvent,
  type AlertRuleRequest,
} from "../../../domain/alerts";

const PATHS = {
  rules: "/v1/alerts/rules",
  rule: (id: string) => `/v1/alerts/rules/${id}`,
  evaluateRule: (id: string) => `/v1/alerts/rules/${id}/evaluate`,
  history: "/v1/alerts/history",
  ruleHistory: (id: string) => `/v1/alerts/rules/${id}/history`,
} as const;

export const alerts = {
  async listRules(organizationId?: string): Promise<AlertRule[]> {
    const response = await restClient.get<{ rules: Parameters<typeof alertRulesFromRaw>[0] }>(PATHS.rules, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return alertRulesFromRaw(response.rules);
  },

  async getRule(id: string, organizationId?: string): Promise<AlertRule> {
    const response = await restClient.get<Parameters<typeof alertRuleFromRaw>[0]>(PATHS.rule(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return alertRuleFromRaw(response);
  },

  async createRule(req: AlertRuleRequest, organizationId?: string): Promise<AlertRule> {
    const response = await restClient.post<Parameters<typeof alertRuleFromRaw>[0]>(PATHS.rules, req, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return alertRuleFromRaw(response);
  },

  async updateRule(id: string, req: AlertRuleRequest, organizationId?: string): Promise<AlertRule> {
    const response = await restClient.patch<Parameters<typeof alertRuleFromRaw>[0]>(PATHS.rule(id), req, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return alertRuleFromRaw(response);
  },

  async deleteRule(id: string, organizationId?: string): Promise<void> {
    await restClient.delete(PATHS.rule(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async evaluateRule(id: string, organizationId?: string): Promise<boolean> {
    const response = await restClient.post<{ transitioned: boolean }>(PATHS.evaluateRule(id), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return response.transitioned;
  },

  async getHistory(ruleId?: string, limit?: number, organizationId?: string): Promise<AlertEvent[]> {
    const path = ruleId ? PATHS.ruleHistory(ruleId) : PATHS.history;
    const response = await restClient.get<{ events: Parameters<typeof alertEventFromRaw>[0][] }>(path, {
      params: {
        limit,
        organization_id: organizationId || getOrganizationContext(),
      },
    });
    return response.events.map(alertEventFromRaw);
  },
};
