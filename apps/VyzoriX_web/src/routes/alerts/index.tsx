import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import {
  ALERT_METRICS,
  ALERT_CONDITIONS,
  type AlertRule,
  type AlertRuleRequest,
} from '@vyzorix/api-client';
import {
  useAlertRules,
  useCreateAlertRule,
  useDeleteAlertRule,
  useEvaluateAlertRule,
} from '@/hooks/alerts/use-alerts';

const METRIC_LABELS: Record<(typeof ALERT_METRICS)[number], string> = {
  device_offline_count: 'Offline devices (count)',
  device_offline_percent: 'Offline devices (%)',
  command_failure_rate: 'Command failure rate (%)',
};

const STATE_COLORS: Record<string, string> = {
  firing: '#ef4444',
  pending: '#f59e0b',
  inactive: '#22c55e',
};

function RuleForm({
  onSubmit,
  disabled,
}: {
  onSubmit: (req: AlertRuleRequest) => void;
  disabled: boolean;
}) {
  const [name, setName] = useState('');
  const [metric, setMetric] = useState<(typeof ALERT_METRICS)[number]>('device_offline_count');
  const [condition, setCondition] = useState<(typeof ALERT_CONDITIONS)[number]>('gt');
  const [threshold, setThreshold] = useState('1');
  const [forSeconds, setForSeconds] = useState('0');
  const [webhookUrl, setWebhookUrl] = useState('');

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit({
          name,
          metric,
          condition,
          threshold: Number(threshold),
          forSeconds: Number(forSeconds),
          webhookUrl: webhookUrl || undefined,
          enabled: true,
        });
      }}
    >
      <div className="grid grid-cols-2 gap-3">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Rule name"
          required
          className="rounded border px-3 py-2"
        />
        <select value={metric} onChange={(e) => setMetric(e.target.value as (typeof ALERT_METRICS)[number])} className="rounded border px-3 py-2">
          {ALERT_METRICS.map((m) => (
            <option key={m} value={m}>{METRIC_LABELS[m]}</option>
          ))}
        </select>
        <select value={condition} onChange={(e) => setCondition(e.target.value as (typeof ALERT_CONDITIONS)[number])} className="rounded border px-3 py-2">
          {ALERT_CONDITIONS.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
        <input
          value={threshold}
          onChange={(e) => setThreshold(e.target.value)}
          type="number"
          step="any"
          placeholder="Threshold"
          required
          className="rounded border px-3 py-2"
        />
        <input
          value={forSeconds}
          onChange={(e) => setForSeconds(e.target.value)}
          type="number"
          min="0"
          placeholder="Pending (seconds)"
          className="rounded border px-3 py-2"
        />
        <input
          value={webhookUrl}
          onChange={(e) => setWebhookUrl(e.target.value)}
          type="url"
          placeholder="Webhook URL (optional)"
          className="rounded border px-3 py-2"
        />
      </div>
      <button
        type="submit"
        disabled={disabled}
        className="mt-4 rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        Create rule
      </button>
    </form>
  );
}

function RuleCard({ rule }: { rule: AlertRule }) {
  const deleteRule = useDeleteAlertRule();
  const evaluateRule = useEvaluateAlertRule();

  return (
    <div className="rounded border p-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold">{rule.name}</h3>
          <p className="text-sm text-muted-foreground">
            {METRIC_LABELS[rule.metric]} {rule.condition} {rule.threshold}
            {rule.forSeconds > 0 ? ` for ${rule.forSeconds}s` : ''}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span
            className="rounded-full px-2 py-1 text-xs font-semibold"
            style={{ backgroundColor: `${STATE_COLORS[rule.state] ?? '#6b7280'}20`, color: STATE_COLORS[rule.state] ?? '#6b7280' }}
          >
            {rule.state}
          </span>
          <button
            onClick={() => evaluateRule.mutate(rule.id)}
            className="rounded border px-2 py-1 text-xs"
            title="Evaluate now"
          >
            Evaluate
          </button>
          <button
            onClick={() => deleteRule.mutate(rule.id)}
            className="rounded border px-2 py-1 text-xs text-red-600"
            title="Delete rule"
          >
            Delete
          </button>
        </div>
      </div>
      {rule.state !== 'inactive' && (
        <p className="mt-2 text-sm">Current value: <strong>{rule.value}</strong></p>
      )}
      {rule.webhookUrl && (
        <p className="mt-1 text-xs text-muted-foreground">Webhook: {rule.webhookUrl}</p>
      )}
    </div>
  );
}

function AlertsPage() {
  const rules = useAlertRules();
  const createRule = useCreateAlertRule();

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-2xl font-bold">Alert Rules</h1>

      <div className="mb-8 rounded border p-4">
        <h2 className="mb-4 font-semibold">New rule</h2>
        <RuleForm onSubmit={(req) => createRule.mutate(req)} disabled={createRule.isPending} />
      </div>

      {rules.isLoading && <p>Loading rules…</p>}
      {rules.isError && <p className="text-red-600">Failed to load rules.</p>}

      <div className="space-y-3">
        {rules.data?.map((rule) => (
          <RuleCard key={rule.id} rule={rule} />
        ))}
        {rules.data?.length === 0 && (
          <p className="text-muted-foreground">No alert rules yet. Create one above.</p>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute('/alerts')({
  component: AlertsPage,
});
