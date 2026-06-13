import React from "react";

import { layers, typography, colors } from "../lib/themes";

interface ActionPanelProps {
  title: string;
  description: string;
  primaryAction: {
    label: string;
    onClick: () => void;
    icon?: React.ReactNode;
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
}

export default function ActionPanel({
  title,
  description,
  primaryAction,
  secondaryAction,
}: ActionPanelProps) {
  return (
    <div
      className={`${layers.card.base} ${layers.card.standard} flex flex-col md:flex-row items-center justify-between gap-6`}
    >
      <div className="flex-1">
        <h3 className={`${typography.heading.h3} ${colors.text.primary} mb-1`}>{title}</h3>
        <p className={`${typography.body.small} ${colors.text.tertiary}`}>{description}</p>
      </div>

      <div className="flex flex-col sm:flex-row items-center gap-3 w-full md:w-auto">
        {secondaryAction && (
          <button
            type="button"
            onClick={secondaryAction.onClick}
            className={`${layers.button.secondary} px-6 w-full sm:w-auto`}
          >
            {secondaryAction.label}
          </button>
        )}
        <button
          type="button"
          onClick={primaryAction.onClick}
          className={`${layers.button.primary} px-6 w-full sm:w-auto`}
        >
          {primaryAction.label}
          {primaryAction.icon}
        </button>
      </div>
    </div>
  );
}
