import { createFileRoute } from "@tanstack/react-router";
import type { ReactElement } from "react";

import { LogConsole } from "@/components/logs/log-console";
import { Card } from "@/components/ui/card";

const LogsPage = (): ReactElement => {
  return (
    <Card className="h-[calc(100vh-10rem)] overflow-hidden p-0">
      <LogConsole height="h-full" />
    </Card>
  );
};

export const Route = createFileRoute("/_app/logs")({
  head: () => ({ meta: [{ title: "Logs — Vyzorix" }] }),
  component: LogsPage,
});
