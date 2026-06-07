import type { ReactElement } from "react";

import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const AppFooter = (): ReactElement => {
  const { serverUrl, deviceId } = useVyzorixConfig();
  return (
    <footer className="border-t bg-background/95 px-4 py-2 text-[11px] text-muted-foreground">
      <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
        <span>Vyzorix Dashboard · phase 1</span>
        <span>
          · device: <code className="font-mono">{deviceId}</code>
        </span>
        <span>
          · server: <code className="font-mono">{serverUrl}</code>
        </span>
        <span className="ml-auto">© {new Date().getFullYear()}</span>
      </div>
    </footer>
  );
};
