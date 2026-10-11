import { useGetUpdatesHistoryPushId } from '@/generated-rq/updates/update-management';

export function usePushDetail(pushId: string | undefined) {
  return useGetUpdatesHistoryPushId(
    pushId ?? '',
    { query: { enabled: pushId !== undefined } },
  );
}
