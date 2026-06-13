import React from "react";

import { typography, colors } from "../lib/themes";

interface DataRowProps {
  label: string;
  value: string | React.ReactNode;
  isMono?: boolean;
}

export default function DataRow({ label, value, isMono = false }: DataRowProps) {
  return (
    <div className="flex justify-between items-center py-2">
      <span className={`${typography.label.micro} ${colors.text.tertiary}`}>{label}</span>
      <span
        className={`${isMono ? typography.data.monospace : typography.data.value} ${colors.text.primary}`}
      >
        {value}
      </span>
    </div>
  );
}
