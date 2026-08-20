import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { alerts, type AlertRuleRequest } from '@vyzorix/api-client';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

const alertsKeys = {
  rules: (orgId: string | null) => ['alerts', 'rules', orgId] as const,
  history: (orgId: string | null, ruleId?: string) => ['alerts', 'history', orgId, ruleId] as const,
};

export function useAlertRules() {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: alertsKeys.rules(organizationId),
    queryFn: () => alerts.listRules(organizationId ?? undefined),
    enabled: organizationId !== null,
    refetchInterval: 30_000,
  });
}

export function useAlertHistory(ruleId?: string, limit?: number) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: alertsKeys.history(organizationId, ruleId),
    queryFn: () => alerts.getHistory(ruleId, limit, organizationId ?? undefined),
    enabled: organizationId !== null,
    refetchInterval: 60_000,
  });
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (req: AlertRuleRequest) => alerts.createRule(req, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useUpdateAlertRule() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: AlertRuleRequest }) =>
      alerts.updateRule(id, req, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useDeleteAlertRule() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => alerts.deleteRule(id, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useEvaluateAlertRule() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => alerts.evaluateRule(id, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}
