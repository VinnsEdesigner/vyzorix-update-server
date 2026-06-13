

import { colors, typography } from "../lib/themes";

interface StatusIndicatorProps {
  status: "active" | "standby" | "offline";
  label: string;
}

export default function StatusIndicator({ status, label }: StatusIndicatorProps) {
  const statusConfig = {
    active: {
      color: colors.accent.base,
      bg: colors.accent.muted,
      border: colors.border.activeSubtle,
      pulse: true,
    },
    standby: {
      color: "bg-amber-500",
      bg: "bg-amber-500/10",
      border: "border-amber-500/20",
      pulse: false,
    },
    offline: {
      color: "bg-slate-500",
      bg: "bg-slate-500/10",
      border: "border-slate-500/20",
      pulse: false,
    },
  };

  const config = statusConfig[status];

  return (
    <div
      className={`inline-flex items-center gap-2.5 px-3 py-1.5 rounded-md border ${config.border} ${config.bg}`}
    >
      <div className="relative flex h-2 w-2 items-center justify-center">
        {config.pulse && (
          <span
            className={`absolute inline-flex h-full w-full rounded-full opacity-75 animate-ping ${config.color.replace("bg-", "bg-")}`}
          ></span>
        )}
        <span className={`relative inline-flex rounded-full h-1.5 w-1.5 ${config.color}`}></span>
      </div>
      <span className={`${typography.label.micro} ${colors.text.primary}`}>{label}</span>
    </div>
  );
}
