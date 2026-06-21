import { createFileRoute, Link } from "@tanstack/react-router";
import { ExternalLink, BookOpen, Code, FlaskConical, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

/**
 * GraphQL Schema Explorer Component
 * Displays the schema types and their fields.
 */
function SchemaExplorer() {
  const { serverUrl } = useVyzorixConfig();

  const schemaTypes = [
    { name: "Query", fields: ["devices", "device", "dashboard", "telemetryHistory", "telemetryStats", "pendingCommands", "connectionStatus", "allConnections"] },
    { name: "Mutation", fields: ["sendCommand", "cancelCommand", "retryCommand", "updateFCMToken", "deleteDevice"] },
    { name: "Device", fields: ["id", "deviceId", "model", "manufacturer", "osVersion", "appVersion", "status", "lastSeen", "createdAt"] },
    { name: "TelemetryFrame", fields: ["timestamp", "riskScore", "thermalTemp", "bufferLevel", "audioMode", "fcmToken"] },
    { name: "Command", fields: ["id", "dispatchId", "deviceId", "command", "status", "createdAt", "updatedAt", "delivery", "result"] },
    { name: "Connection", fields: ["deviceId", "status", "connectedAt", "lastActivity", "ipAddress", "userAgent"] },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <BookOpen className="h-4 w-4" />
          Schema Reference
        </CardTitle>
        <CardDescription>Quick reference for GraphQL types and fields</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {schemaTypes.map((type) => (
          <div key={type.name}>
            <h4 className="text-sm font-medium text-primary">{type.name}</h4>
            <div className="mt-1 flex flex-wrap gap-1">
              {type.fields.map((field) => (
                <code
                  key={field}
                  className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground"
                >
                  {field}
                </code>
              ))}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

/**
 * Example Queries Component
 */
function ExampleQueries() {
  const copyToClipboard = async (query: string) => {
    await navigator.clipboard.writeText(query);
  };

  const examples = [
    {
      title: "Get All Devices",
      query: `query {
  devices(limit: 10) {
    id
    deviceId
    model
    status
    lastSeen
  }
}`,
    },
    {
      title: "Get Device Detail",
      query: `query {
  device(id: "device-123") {
    id
    deviceId
    model
    manufacturer
    telemetry {
      riskScore
      thermalTemp
      bufferLevel
    }
    commands(last: 5) {
      id
      command
      status
    }
  }
}`,
    },
    {
      title: "Get Dashboard Data",
      query: `query {
  dashboard(limit: 10) {
    devices {
      id
      deviceId
      status
    }
    connections {
      deviceId
      status
    }
    totalDevices
    onlineDevices
    totalCommands
    pendingCommands
  }
}`,
    },
    {
      title: "Send Command",
      query: `mutation {
  sendCommand(input: {
    deviceId: "device-123"
    command: "RESTART_APP"
  }) {
    dispatchId
    status
    command
  }
}`,
    },
    {
      title: "Get Telemetry History",
      query: `query {
  telemetryHistory(
    deviceId: "device-123"
    startTime: "2024-01-01T00:00:00Z"
    endTime: "2024-01-02T00:00:00Z"
    limit: 100
  ) {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
  }
}`,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Code className="h-4 w-4" />
          Example Queries
        </CardTitle>
        <CardDescription>Click to copy GraphQL queries</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {examples.map((example) => (
          <div key={example.title}>
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-medium">{example.title}</h4>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => copyToClipboard(example.query)}
              >
                Copy
              </Button>
            </div>
            <pre className="mt-1 max-h-40 overflow-auto rounded-md border bg-muted/50 p-2 text-xs">
              {example.query}
            </pre>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

/**
 * GraphQL Playground Page
 * Provides links and information about the GraphQL playground.
 */
function PlaygroundPage(): React.ReactElement {
  const { serverUrl } = useVyzorixConfig();
  const playgroundUrl = `${serverUrl}/playground`;
  const graphqlUrl = `${serverUrl}/graphql`;
  const voyagerUrl = `${serverUrl}/voyager`; // Optional: GraphQL Voyager for schema visualization

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Dashboard
          </Button>
        </Link>
      </div>

      <div className="space-y-1">
        <h1 className="text-3xl font-bold tracking-tight">GraphQL API</h1>
        <p className="text-muted-foreground">
          Interactive GraphQL playground and API documentation for Vyzorix.
        </p>
      </div>

      {/* URLs */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">API Endpoints</CardTitle>
          <CardDescription>GraphQL API URLs for your server</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between rounded-md border p-3">
            <div>
              <p className="text-sm font-medium">GraphQL Endpoint</p>
              <code className="text-xs text-muted-foreground">{graphqlUrl}</code>
            </div>
          </div>
          <div className="flex items-center justify-between rounded-md border p-3">
            <div>
              <p className="text-sm font-medium">Playground</p>
              <code className="text-xs text-muted-foreground">{playgroundUrl}</code>
            </div>
            <Button asChild size="sm">
              <a href={playgroundUrl} target="_blank" rel="noreferrer">
                <ExternalLink className="mr-2 h-4 w-4" />
                Open
              </a>
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Quick Links */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <FlaskConical className="h-4 w-4" />
              Getting Started
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>1. Open the Playground using the link above</p>
            <p>2. The playground automatically includes your session cookie</p>
            <p>3. Try the example queries in the Examples tab</p>
            <p>4. Use the Docs tab to explore the full schema</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <BookOpen className="h-4 w-4" />
              Authentication
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <p>GraphQL uses the same authentication as the REST API:</p>
            <ul className="list-inside list-disc space-y-1 text-muted-foreground">
              <li>Session cookie (recommended for browser)</li>
              <li>Authorization header with Bearer token</li>
            </ul>
            <p className="text-muted-foreground">
              Login at <code className="text-xs">/login</code> to get a session cookie.
            </p>
          </CardContent>
        </Card>
      </div>

      <Separator />

      {/* Schema & Examples */}
      <div className="grid gap-6 lg:grid-cols-2">
        <SchemaExplorer />
        <ExampleQueries />
      </div>

      {/* Migration Guide Link */}
      <Card className="border-dashed">
        <CardContent className="flex items-center justify-between py-6">
          <div>
            <p className="font-medium">Migrating from REST to GraphQL?</p>
            <p className="text-sm text-muted-foreground">
              Check out our migration guide for REST to GraphQL conversion.
            </p>
          </div>
          <Button variant="outline" asChild>
            <Link to="/docs/migration-graphql">
              View Migration Guide
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute("/playground")({
  head: () => ({ meta: [{ title: "GraphQL Playground — Vyzorix" }] }),
  component: PlaygroundPage,
});
