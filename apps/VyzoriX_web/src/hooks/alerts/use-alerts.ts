import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getAlerts, type AlertRuleRequest } from '@vyzorix/api-client';

const alertsKeys = {
  rules: (orgId: string | null) => ['alerts', 'rules', orgId] as const,
  history: (orgId: string | null, ruleId?: string) => ['alerts', 'history', orgId, ruleId] as const,
};

export function useAlertRules() {
  return useQuery({
    queryKey: alertsKeys.rules(null),
    queryFn: () => getAlerts().getAlertsRules(),
    enabled: true,
    refetchInterval: 30_000,
  });
}

export function useAlertHistory(ruleId?: string, limit?: number) {
  return useQuery({
    queryKey: alertsKeys.history(null, ruleId),
    queryFn: () => getAlerts().getAlertsRulesIdHistory(ruleId ?? '', { limit }),
    enabled: ruleId !== undefined,
    refetchInterval: 60_000,
  });
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: AlertRuleRequest) => getAlerts().postAlertsRules(req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useUpdateAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: AlertRuleRequest }) =>
      getAlerts().patchAlertsRulesId(id, req),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useDeleteAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => getAlerts().deleteAlertsRulesId(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}

export function useEvaluateAlertRule() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => getAlerts().postAlertsRulesIdEvaluate(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }),
  });
}
