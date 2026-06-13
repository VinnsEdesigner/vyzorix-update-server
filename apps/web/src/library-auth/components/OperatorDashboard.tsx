

import DataCard from "./DataCard";
import DataRow from "./DataRow";
import StatusIndicator from "./StatusIndicator";
import MetricsGrid from "./MetricsGrid";
import ActionPanel from "./ActionPanel";

export default function OperatorDashboard() {
  const systemMetrics = [
    { label: "Active Sessions", value: "1,492", trend: "up" as const, trendValue: "4.2%" },
    { label: "Latency (ms)", value: "12.4", trend: "down" as const, trendValue: "1.2%" },
    { label: "Database Calls", value: "45.1K" },
    { label: "Uptime", value: "99.99%", trend: "neutral" as const, trendValue: "Stable" },
  ];

  return (
    <div className="space-y-6 animate-fade-in text-slate-100">
      <div className="mb-8 border-b border-white/5 pb-4">
        <h1 className="text-2xl font-semibold tracking-tight text-white">
          Infrastructure Overview
        </h1>
        <p className="text-slate-400 text-sm mt-1">Real-time telemetrics and component status</p>
      </div>

      <MetricsGrid metrics={systemMetrics} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <ActionPanel
            title="Update Protocol Certificates"
            description="Recent security advisories require rotation of backend access certificates. Ensure systems remain fully compliant."
            primaryAction={{
              label: "Rotate Certificates",
              onClick: () => console.log("Rotating..."),
            }}
            secondaryAction={{
              label: "View Logs",
              onClick: () => console.log("Viewing..."),
            }}
          />

          <DataCard title="Active Node Activity" badge="Live">
            <DataRow
              label="Region: Paris, FR"
              value={<StatusIndicator status="active" label="Operational" />}
            />
            <DataRow
              label="Region: Frankfurt, DE"
              value={<StatusIndicator status="active" label="Operational" />}
            />
            <DataRow
              label="Region: Singapore, SG"
              value={<StatusIndicator status="standby" label="Standby" />}
            />
            <DataRow
              label="Region: Tokyo, JP"
              value={<StatusIndicator status="offline" label="Maintenance" />}
            />
          </DataCard>
        </div>

        <div className="space-y-6">
          <DataCard title="Connection Identity" badge="Verified">
            <DataRow label="Operator ID" value="VXZ-64981" isMono />
            <DataRow label="Clearance" value="Level 4" />
            <DataRow label="Auth Method" value="Biometric + OIDC" />
            <DataRow label="Token Lifecycle" value="Expires in 42m" />
          </DataCard>
        </div>
      </div>
    </div>
  );
}
