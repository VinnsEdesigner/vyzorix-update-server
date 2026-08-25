import { useGetUpdatesHistoryPushId } from '@/generated-rq/updates/update-management';

export function usePushDetail(pushId: string | undefined) {
  return useGetUpdatesHistoryPushId(
    pushId ?? '',
    { query: { queryKey: ['updates-push-detail', pushId] as const, enabled: pushId !== undefined } },
  );
}
