// Thin wrappers over the generated TanStack Query hooks for alerts.
// Cache invalidation is automatic: the generated mutations invalidate the
// org-scoped keys their `mutationInvalidates` tag declares
// so no manual invalidateQueries is needed.
import {
  useGetAlertsRules,
  useGetAlertsRulesIdHistory,
  usePostAlertsRules,
  usePatchAlertsRulesId,
  useDeleteAlertsRulesId,
  usePostAlertsRulesIdEvaluate,
} from '@/generated-rq/alerts/alert-rules';

export function useAlertRules() {
  return useGetAlertsRules({ query: { refetchInterval: 30_000 } });
}

export function useAlertHistory(ruleId?: string, limit?: number) {
  return useGetAlertsRulesIdHistory(
    ruleId ?? '',
    { limit },
    { query: { enabled: ruleId !== undefined, refetchInterval: 60_000 } },
  );
}

export function useCreateAlertRule() {
  return usePostAlertsRules();
}

export function useUpdateAlertRule() {
  return usePatchAlertsRulesId();
}

export function useDeleteAlertRule() {
  return useDeleteAlertsRulesId();
}

export function useEvaluateAlertRule() {
  return usePostAlertsRulesIdEvaluate();
}
