import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  registration,
  type CreateInboxRequest,
  type CreateInboxResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useRegisterDevice() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (request: CreateInboxRequest) =>
      registration.createInboxRequest(request, organizationId ?? undefined),
    onSuccess: (_, request) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
      queryClient.invalidateQueries({
        queryKey: queryKeys.registrationInboxEntry(request.imei),
      });
    },
  });
}

export type { CreateInboxRequest, CreateInboxResult };
