import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getInbox, type CreateInboxRequest, type InboxEntryResponse } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useRegisterDevice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateInboxRequest) =>
      getInbox().postDeviceInbox(request),
    onSuccess: (_, request) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.invalidateQueries({
        queryKey: queryKeys.registrationInboxEntry(request.imei),
      });
    },
  });
}

export type { CreateInboxRequest, InboxEntryResponse };
