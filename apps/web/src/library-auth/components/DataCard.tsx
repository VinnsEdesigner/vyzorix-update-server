import React from "react";

import { layers, typography, colors } from "../lib/themes";

interface DataCardProps {
  title: string;
  badge?: string;
  children: React.ReactNode;
  action?: React.ReactNode;
}

export default function DataCard({ title, badge, children, action }: DataCardProps) {
  return (
    <div className={`${layers.card.base} ${layers.card.subtle}`}>
      <div className="flex items-center justify-between mb-5 border-b border-white/5 pb-4">
        <div className="flex items-center gap-3">
          <h3 className={`${typography.heading.h3} ${colors.text.primary}`}>{title}</h3>
          {badge && (
            <span
              className={`${typography.label.micro} ${colors.text.primary} bg-white/10 px-2 py-0.5 rounded`}
            >
              {badge}
            </span>
          )}
        </div>
        {action && <div>{action}</div>}
      </div>
      <div className="space-y-4">{children}</div>
    </div>
  );
}
