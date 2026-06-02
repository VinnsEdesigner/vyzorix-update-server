import { createFileRoute } from "@tanstack/react-router";
import { Card } from "@/components/ui/card";
import { LogConsole } from "@/components/logs/log-console";

export const Route = createFileRoute("/_app/logs")({
  head: () => ({ meta: [{ title: "Logs — Vyzorix" }] }),
  component: LogsPage,
});

function LogsPage() {
  return (
    <Card className="h-[calc(100vh-10rem)] overflow-hidden p-0">
      <LogConsole height="h-full" />
    </Card>
  );
}