import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import {
  CONTACT_POINT_CHANNELS,
  type ContactPoint,
  type ContactPointChannel,
  type ContactPointRequest,
} from '@vyzorix/api-client';
import {
  useContactPoints,
  useCreateContactPoint,
  useDeleteContactPoint,
  useTestContactPoint,
} from '@/hooks/notifications/use-contact-points';

const CHANNEL_LABELS: Record<ContactPointChannel, string> = {
  email: 'Email',
  webhook: 'Webhook',
  slack: 'Slack',
};

const CHANNEL_COLORS: Record<string, string> = {
  email: '#3b82f6',
  webhook: '#8b5cf6',
  slack: '#059669',
};

function ContactPointForm({
  onSubmit,
  disabled,
}: {
  onSubmit: (req: ContactPointRequest) => void;
  disabled: boolean;
}) {
  const [name, setName] = useState('');
  const [channel, setChannel] = useState<ContactPointChannel>('webhook');
  const [secret, setSecret] = useState('');
  const [config, setConfig] = useState('');
  const [enabled, setEnabled] = useState(true);

  const configFromText = (text: string): Record<string, string> => {
    const obj: Record<string, string> = {};
    text.split('\n').forEach((line) => {
      const [key, ...rest] = line.split('=');
      if (key && rest.length > 0) obj[key.trim()] = rest.join('=').trim();
    });
    return obj;
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit({
          name,
          channel,
          secret: secret || undefined,
          config: configFromText(config),
          enabled,
        });
      }}
    >
      <div className="grid grid-cols-2 gap-3">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Contact point name"
          required
          className="rounded border px-3 py-2"
        />
        <select
          value={channel}
          onChange={(e) => setChannel(e.target.value as ContactPointChannel)}
          className="rounded border px-3 py-2"
        >
          {CONTACT_POINT_CHANNELS.map((c) => (
            <option key={c} value={c}>
              {CHANNEL_LABELS[c]}
            </option>
          ))}
        </select>
        <input
          value={secret}
          onChange={(e) => setSecret(e.target.value)}
          type="password"
          placeholder="Webhook secret (optional)"
          className="rounded border px-3 py-2"
        />
        <label className="flex items-center gap-2 px-3 py-2">
          <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />
          <span className="text-sm">Enabled</span>
        </label>
      </div>
      <textarea
        value={config}
        onChange={(e) => setConfig(e.target.value)}
        placeholder="Config (one per line: to=ops@example.com, url=https://hooks.slack.com/...)"
        rows={4}
        className="mt-3 w-full rounded border px-3 py-2"
      />
      <button
        type="submit"
        disabled={disabled}
        className="mt-4 rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        Create contact point
      </button>
    </form>
  );
}

function ContactPointCard({ point }: { point: ContactPoint }) {
  const deletePoint = useDeleteContactPoint();
  const testPoint = useTestContactPoint();

  return (
    <div className="rounded border p-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold">{point.name}</h3>
          <p className="text-sm text-muted-foreground">
            {CHANNEL_LABELS[point.channel as keyof typeof CHANNEL_LABELS] ?? point.channel}
            {point.secret && ' (signed)'}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span
            className="rounded-full px-2 py-1 text-xs font-semibold"
            style={{
              backgroundColor: `${CHANNEL_COLORS[point.channel ?? ''] ?? '#6b7280'}20`,
              color: CHANNEL_COLORS[point.channel ?? ''] ?? '#6b7280',
            }}
          >
            {point.channel}
          </span>
          <button
            onClick={() => point.id && testPoint.mutate({ id: point.id })}
            className="rounded border px-2 py-1 text-xs"
            title="Send test notification"
          >
            Test
          </button>
          <button
            onClick={() => point.id && deletePoint.mutate({ id: point.id })}
            className="rounded border px-2 py-1 text-xs text-red-600"
            title="Delete"
          >
            Delete
          </button>
        </div>
      </div>
      <div className="mt-2 text-xs text-muted-foreground">
        {Object.entries(point.config ?? {}).map(([k, v]) => (
          <div key={k}>
            <strong>{k}:</strong> {v.length > 40 ? v.slice(0, 40) + '…' : v}
          </div>
        ))}
      </div>
    </div>
  );
}

function NotificationsPage() {
  const points = useContactPoints();
  const createPoint = useCreateContactPoint();

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-2xl font-bold">Contact Points</h1>

      <div className="mb-8 rounded border p-4">
        <h2 className="mb-4 font-semibold">New contact point</h2>
        <ContactPointForm onSubmit={(req) => createPoint.mutate({ data: req })} disabled={createPoint.isPending} />
      </div>

      {points.isLoading && <p>Loading contact points…</p>}
      {points.isError && <p className="text-red-600">Failed to load contact points.</p>}

      <div className="space-y-3">
        {(points.data?.contact_points ?? []).map((point) => (
          <ContactPointCard key={point.id} point={point} />
        ))}
        {(points.data?.contact_points ?? []).length === 0 && (
          <p className="text-muted-foreground">No contact points yet. Create one above.</p>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute('/notifications')({
  component: NotificationsPage,
});
