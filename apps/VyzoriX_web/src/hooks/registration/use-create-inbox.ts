import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  registration,
  type CreateInboxRequest,
  type CreateInboxResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useCreateInboxRequest() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (request: CreateInboxRequest) =>
      registration.createInboxRequest(request, organizationId ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.registrationInbox() });
    },
  });
}

export type { CreateInboxRequest, CreateInboxResult };
