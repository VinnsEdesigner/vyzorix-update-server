

import { layers, typography, colors } from "../lib/themes";

interface MetricsGridProps {
  metrics: Array<{
    label: string;
    value: string;
    trend?: "up" | "down" | "neutral";
    trendValue?: string;
  }>;
}

export default function MetricsGrid({ metrics }: MetricsGridProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {metrics.map((metric, idx) => (
        <div key={idx} className={`${layers.card.base} ${layers.card.subtle} py-4`}>
          <span className={`${typography.label.micro} ${colors.text.tertiary} block mb-2`}>
            {metric.label}
          </span>
          <div className="flex items-end justify-between">
            <span className={`${typography.heading.h2} ${colors.text.primary}`}>
              {metric.value}
            </span>
            {metric.trend && metric.trendValue && (
              <span
                className={`text-xs font-semibold tracking-wide ${
                  metric.trend === "up"
                    ? "text-emerald-400"
                    : metric.trend === "down"
                      ? "text-rose-400"
                      : colors.text.tertiary
                }`}
              >
                {metric.trend === "up" ? "↑" : metric.trend === "down" ? "↓" : "−"}{" "}
                {metric.trendValue}
              </span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
