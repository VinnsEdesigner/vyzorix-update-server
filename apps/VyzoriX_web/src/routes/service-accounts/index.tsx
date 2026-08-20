import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import {
  serviceAccounts as api,
  type ServiceAccount,
} from '@vyzorix/api-client';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

const saKeys = {
  list: (orgId: string | null) => ['service-accounts', 'list', orgId] as const,
  tokens: (orgId: string | null, saId: string) => ['service-accounts', 'tokens', orgId, saId] as const,
};

function useServiceAccounts() {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: saKeys.list(organizationId),
    queryFn: () => api.list(organizationId ?? undefined),
    enabled: organizationId !== null,
    refetchInterval: 60_000,
  });
}

function ServiceAccountCard({ sa }: { sa: ServiceAccount }) {
  const organizationId = useCurrentOrganizationId();
  const queryClient = useQueryClient();
  const deleteAccount = useMutation({
    mutationFn: () => api.delete(sa.id, organizationId ?? undefined),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['service-accounts'] }),
  });
  const rotateToken = useMutation({
    mutationFn: () => api.rotateToken(sa.id, 'any-token-id', organizationId ?? undefined),
  });

  return (
    <div className="rounded border p-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold">{sa.name}</h3>
          <p className="text-sm text-muted-foreground">ID: {sa.id.slice(0, 8)}…</p>
        </div>
        <div className="flex items-center gap-2">
          <span
            className="rounded-full px-2 py-1 text-xs font-semibold"
            style={{ backgroundColor: `${sa.enabled ? '#22c55e' : '#6b7280'}20`, color: sa.enabled ? '#22c55e' : '#6b7280' }}
          >
            {sa.enabled ? 'active' : 'disabled'}
          </span>
          <button
            onClick={() => rotateToken.mutate()}
            className="rounded border px-2 py-1 text-xs"
            title="Rotate a token"
          >
            Rotate
          </button>
          <button
            onClick={() => deleteAccount.mutate()}
            className="rounded border px-2 py-1 text-xs text-red-600"
            title="Delete"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}

function ServiceAccountsPage() {
  const organizationId = useCurrentOrganizationId();
  const queryClient = useQueryClient();
  const accounts = useServiceAccounts();
  const [name, setName] = useState('');

  const createAccount = useMutation({
    mutationFn: () => api.create(name, organizationId ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['service-accounts'] });
      setName('');
    },
  });

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-2xl font-bold">Service Accounts</h1>

      <div className="mb-8 rounded border p-4">
        <h2 className="mb-4 font-semibold">New service account</h2>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            createAccount.mutate();
          }}
          className="flex gap-2"
        >
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Account name (e.g., CI deployer)"
            required
            className="flex-1 rounded border px-3 py-2"
          />
          <button
            type="submit"
            disabled={createAccount.isPending}
            className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
          >
            Create
          </button>
        </form>
      </div>

      {accounts.isLoading && <p>Loading service accounts…</p>}
      {accounts.isError && <p className="text-red-600">Failed to load service accounts.</p>}

      <div className="space-y-3">
        {accounts.data?.map((sa) => (
          <ServiceAccountCard key={sa.id} sa={sa} />
        ))}
        {accounts.data?.length === 0 && (
          <p className="text-muted-foreground">No service accounts yet. Create one above.</p>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute('/service-accounts')({
  component: ServiceAccountsPage,
});
