// Re-export generated TanStack Query hooks for alerts.
// GETs use generated useQuery hooks, mutations use generated useMutation
// hooks — both properly typed and auto-keyed by orval.
import { useQueryClient } from '@tanstack/react-query';
import {
  useGetAlertsRules,
  useGetAlertsRulesIdHistory,
  usePostAlertsRules,
  usePatchAlertsRulesId,
  useDeleteAlertsRulesId,
  usePostAlertsRulesIdEvaluate,
} from '@/generated-rq/alerts/alert-rules';

export function useAlertRules() {
  return useGetAlertsRules({ query: { queryKey: ['alerts', 'rules'] as const, refetchInterval: 30_000 } });
}

export function useAlertHistory(ruleId?: string, limit?: number) {
  return useGetAlertsRulesIdHistory(
    ruleId ?? '',
    { limit },
    { query: { queryKey: ['alerts', 'history', ruleId] as const, enabled: ruleId !== undefined, refetchInterval: 60_000 } },
  );
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient();
  return usePostAlertsRules({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }) },
  });
}

export function useUpdateAlertRule() {
  const queryClient = useQueryClient();
  return usePatchAlertsRulesId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }) },
  });
}

export function useDeleteAlertRule() {
  const queryClient = useQueryClient();
  return useDeleteAlertsRulesId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }) },
  });
}

export function useEvaluateAlertRule() {
  const queryClient = useQueryClient();
  return usePostAlertsRulesIdEvaluate({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alerts'] }) },
  });
}
