import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getInbox, type CreateInboxRequest, type InboxEntryResponse } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useCreateInboxRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateInboxRequest) =>
      getInbox().postDeviceInbox(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
    },
  });
}

export type { CreateInboxRequest, InboxEntryResponse };
