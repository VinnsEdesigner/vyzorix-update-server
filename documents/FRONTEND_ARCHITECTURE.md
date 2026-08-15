# Frontend Architecture Foundation

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-22  
> **Target:** Production MVP  

---

## Table of Contents

1. [Philosophy](#1-philosophy)
2. [Architecture Overview](#2-architecture-overview)
3. [Layer Definitions](#3-layer-definitions)
4. [Dependency Rules](#4-dependency-rules)
5. [Directory Structure](#5-directory-structure)
6. [Domain Layer](#6-domain-layer)
7. [Data Layer](#7-data-layer)
8. [Presentation Layer](#8-presentation-layer)
9. [UI Layer](#9-ui-layer)
10. [Implementation Guide](#10-implementation-guide)
11. [Debugging Guide](#11-debugging-guide)
12. [Migration Checklist](#12-migration-checklist)

---

## 1. Philosophy

### 1.1 Why Layered Architecture?

| Approach | Pros | Cons |
|----------|------|------|
| **Flat (Current)** | Quick to start | Hard to maintain, hard to debug |
| **Layered** | Clear boundaries, easy to test, predictable | Initial setup time |

### 1.2 Core Principles

1. **Single Responsibility**: Each layer has ONE job
2. **Dependency Rule**: Dependencies flow inward only
3. **Testability**: Every layer can be tested in isolation
4. **Predictability**: You always know where to look

### 1.3 The Problem We're Solving

```
FLAT ARCHITECTURE:

                                                                     
  Component  API  Context  Hook  
                                                                     
  Problem:                                                         
  - Where do I put this logic?                                      
  - Where is this data coming from?                                 
  - Why is this breaking?                                           
  - How do I test this?                                             
  - Can I reuse this?                                               
                                                                     


LAYERED ARCHITECTURE:

                                                                     
  UI Layer  Presentation Layer  Domain  
                                                                    
                                                                    
                                                                    
                              Data Layer                             
                                                                     
  Benefits:                                                         
  - I always know where to look                                     
  - I know exactly where something broke                            
  - I can test each layer independently                             
  - I can reuse any layer                                          
                                                                     

```

---

## 2. Architecture Overview

### 2.1 The Four Layers

```

                                                                     
                         FRONTEND ARCHITECTURE                       
                                                                     
     
                        UI LAYER                                  
                     (src/ui/)                                   
                                                                  
      Pages, Components, Shared UI                               
      ONLY renders UI. Uses hooks for everything.                 
      NEVER imports from Data or Domain.                          
     
                                                                     
                               uses                                  
                                                                     
     
                     PRESENTATION LAYER                           
                     (src/hooks/)                               
                                                                  
      Custom hooks that:                                         
      - Handle UI logic                                           
      - Transform data for UI                                     
      - Manage state                                              
      NEVER renders UI. NEVER imports from UI layer.             
     
                                                                     
                               uses                                  
                                                                     
     
                        DOMAIN LAYER                             
                     (packages/API_Client/src/domain/)                               
                                                                  
      Pure functions that:                                       
      - Define types and interfaces                             
      - Transform data (no side effects)                        
      - Validate input                                           
      NEVER imports from UI, Presentation, or Data.              
     
                                                                     
                               uses                                  
                                                                     
     
                         DATA LAYER                              
                     (packages/API_Client/src/vyzorServer/)                              
                                                                  
      API clients that:                                          
      - Make HTTP requests                                       
      - Handle authentication                                     
      - Parse responses                                          
      NEVER imports from UI or Presentation.                     
     
                                                                     

```

### 2.2 Data Flow

```

                                                                     
  USER INTERACTS                                                    
                                                                    
                                                                    
     
    UI LAYER                                                        
    - User clicks button                                            
    - Component calls hook function                                 
     
                                                                    
                                                                    
     
    PRESENTATION LAYER                                             
    - Hook validates input                                          
    - Hook calls domain transform                                  
    - Hook calls data layer                                        
    - Hook transforms response                                     
    - Hook returns data to UI                                      
     
                                                                    
                                                                    
     
    DOMAIN LAYER                                                  
    - Validates input                                              
    - Transforms data (pure function)                             
    - Returns transformed data                                    
     
                                                                    
                                                                    
     
    DATA LAYER                                                    
    - Makes API request                                           
    - Returns raw response                                        
     
                                                                     

```

---

## 3. Layer Definitions

### 3.1 UI Layer

| Aspect | Definition |
|--------|-------------|
| **Location** | `src/ui/` |
| **Contains** | Pages, Components, Shared UI |
| **Responsibility** | Render UI, handle user input |
| **Imports From** | ONLY `src/hooks/` |
| **Examples** | `dashboard/page.tsx`, `Button`, `Section` |

### 3.2 Presentation Layer

| Aspect | Definition |
|--------|-------------|
| **Location** | `src/hooks/` |
| **Contains** | Custom React hooks |
| **Responsibility** | UI logic, data transformation, state management |
| **Imports From** | `packages/API_Client/src/domain/`, `packages/API_Client/src/vyzorServer/` |
| **Examples** | `useDeviceList`, `useSendCommand`, `useExport` |

### 3.3 Domain Layer

| Aspect | Definition |
|--------|-------------|
| **Location** | `packages/API_Client/src/domain/` |
| **Contains** | Types, interfaces, pure functions |
| **Responsibility** | Define types, transform data, validate input |
| **Imports From** | NOTHING (no dependencies) |
| **Examples** | `Device` type, `deviceFromAPI()`, `validateImei()` |

### 3.4 Data Layer

| Aspect | Definition |
|--------|-------------|
| **Location** | `packages/API_Client/src/vyzorServer/` |
| **Contains** | API clients, queries, mutations |
| **Responsibility** | Make HTTP requests, handle responses |
| **Imports From** | `packages/API_Client/src/domain/` (types only) |
| **Examples** | GraphQL client, REST client, API endpoints |

---

## 4. Dependency Rules

### 4.1 The Golden Rule

```

                                                                     
   DEPENDENCIES MUST FLOW INWARD ONLY                               
                                                                     
   UI  Presentation  Domain  
                                                                     
      
                                                                     
   NEVER:                                                           
   - UI imports from Domain directly                                
   - UI imports from Data directly                                   
   - Presentation imports from UI                                     
   - Domain imports from anything                                     
   - Data imports from Presentation                                  
   - Data imports from UI                                           
                                                                     

```

### 4.2 Dependency Matrix

| From \ To | UI | Presentation | Domain | Data |
|-----------|-----|--------------|--------|------|
| **UI** |  |  |  |  |
| **Presentation** |  |  |  |  |
| **Domain** |  |  |  |  |
| **Data** |  |  |  (types) |  |

### 4.3 ESLint Rules

```typescript
// src/.eslintrc.js
module.exports = {
  rules: {
    // Prevent imports from upper layers
    'no-restricted-imports': [
      'error',
      {
        patterns: [
          // UI cannot import from these
          { group: ['@vyzorix/api-client/domain/**'], target: 'UI Layer' },
          { group: ['@vyzorix/api-client/vyzorServer/**'], target: 'UI Layer' },
          // Presentation cannot import from UI
          { group: ['@/ui/**'], target: 'Presentation Layer' },
          // Domain cannot import from anything
          { group: ['@/ui/**'], target: 'Domain Layer' },
          { group: ['@/hooks/**'], target: 'Domain Layer' },
          { group: ['@vyzorix/api-client/vyzorServer/**'], target: 'Domain Layer' },
          // Data cannot import from UI or Presentation
          { group: ['@/ui/**'], target: 'Data Layer' },
          { group: ['@/hooks/**'], target: 'Data Layer' },
        ],
      },
    ],
  },
};
```

---

## 5. Directory Structure

### 5.1 Full Structure

```
src/

 ui/                              # UI LAYER
    pages/                       # Route pages
       dashboard/
          dashboard.tsx            # Main dashboard page
          dashboard-overview.tsx
          dashboard-metrics.tsx
          dashboard-commands.tsx
          dashboard-logs.tsx
       commands/
          commands.tsx
          commands-send.tsx
          commands-pending.tsx
          commands-history.tsx
       logs/
          logs.tsx
       device/
          device.tsx
          device-overview.tsx
          device-telemetry.tsx
          device-commands.tsx
          device-history.tsx
       diagnostics/
          diagnostics.tsx
          diagnostics-inspector.tsx
          diagnostics-timeline.tsx
       alerts/
          alerts.tsx
          alerts-active.tsx
          alerts-history.tsx
       updates/
           updates.tsx
   
    components/                  # Reusable components
       ui/                     # Base UI components
          button.tsx
          badge.tsx
          section.tsx
          table.tsx
          tabs.tsx
          input.tsx
          select.tsx
          dropdown-menu.tsx
          dialog.tsx
          skeleton.tsx
          toast.tsx
          tooltip.tsx
          pagination.tsx
          search-input.tsx
          empty-state.tsx
          loading-spinner.tsx
          error-state.tsx
      
       layout/                 # Layout components
          page-header.tsx
          page-tabs.tsx
          sidebar.tsx
          app-layout.tsx
      
       shared/                 # Shared feature components
           section/
              section-container.tsx
              section-header.tsx
              section-content.tsx
           metric-card/
              metric-card-display.tsx
              metric-card-grid.tsx
              metric-progress.tsx
           connection-status/
              connection-status.tsx
              connection-indicator.tsx
           device-selector/
              device-selector.tsx
           command-button/
              command-button.tsx
           command-row/
              command-row.tsx
           log-entry/
              log-entry.tsx
           export-menu/
               export-menu.tsx
               export-scope.tsx
               export-format.tsx
   
    index.ts                    # UI layer exports

 hooks/                          # PRESENTATION LAYER
    auth/
       use-auth.ts
       use-signed-api.ts
    device/
       use-device-stream.ts
       use-device.ts
       use-devices.ts
       use-device-telemetry.ts
       use-device-inspection.ts
       use-device-timeline.ts
    commands/
       use-commands.ts
       use-send-command.ts
       use-cancel-command.ts
       use-command-history.ts
    logs/
       use-logs.ts
       use-log-stream.ts
       use-log-filter.ts
    alerts/
       use-alerts.ts
       use-alert.ts
       use-dismiss-alert.ts
    telemetry/
       use-telemetry.ts
       use-telemetry-stats.ts
    export/
       use-export.ts
       use-export-job.ts
       use-export-dialog.ts
    _shared/
       use-pagination.ts
       use-search.ts
       use-device-selector.ts
       use-time-range.ts
    use-mobile.tsx
    use-operator.ts
    use-server-health.ts

 domain/                         # DOMAIN LAYER
    _shared/                   # SHARED - cross-feature types/functions
       domain-pagination.ts   # Pagination types & helpers
       domain-errors.ts       # Domain error classes
       domain-date.ts         # Date formatting utilities
   
    device/
       device-entity.ts       # Device, DeviceStatus types
       device-mappers.ts      # deviceFromGraphQL(), etc.
       device-validators.ts   # validateDeviceId(), etc.
   
    commands/
       command-entity.ts      # Command, CommandStatus types
       command-constants.ts   # PRESET_COMMANDS array
       command-mappers.ts     # commandFromAPI(), etc.
       command-validators.ts  # validateCommand(), etc.
   
    logs/
       log-entity.ts          # LogEntry, LogType types
       log-mappers.ts         # logFromAPI(), etc.
       log-filters.ts         # filterByType(), etc.
       log-formatters.ts      # formatTimestamp(), etc.
   
    telemetry/
       telemetry-entity.ts    # TelemetryFrame, MetricStats types
       telemetry-mappers.ts   # telemetryFromAPI(), etc.
       telemetry-aggregation.ts # calculateStats(), etc.
   
    alerts/
       alert-entity.ts        # Alert, AlertSeverity types
       alert-mappers.ts       # alertFromAPI(), etc.
       alert-priority.ts      # calculatePriority(), etc.
   
    export/
       export-entity.ts       # ExportFormat, ExportScope types
       export-formatters.ts   # toCSV(), toJSON(), toXML()
       export-filename.ts     # generateFilename()
       export-validators.ts  # validateExportOptions()
   
    index.ts                   # Domain exports (barrel)

 lib/
     api/                        # DATA LAYER
         _shared/              # SHARED - cross-feature data utilities
            graphql-client.ts   # GraphQL client setup
            graphql-fragments.ts # Reusable fragments
            rest-client.ts     # REST client setup
            rest-interceptors.ts # Auth, error handling
            rest-adapters.ts  # graphql→rest adapters
        
         graphql/
            device/
               graphql-device-queries.ts
               graphql-device-mutations.ts
               graphql-device-types.ts  # Raw response types
            commands/
               graphql-command-queries.ts
               graphql-command-mutations.ts
               graphql-command-types.ts
            logs/
               graphql-log-queries.ts
               graphql-log-subscriptions.ts
               graphql-log-types.ts
            telemetry/
               graphql-telemetry-queries.ts
               graphql-telemetry-types.ts
            alerts/
               graphql-alert-queries.ts
               graphql-alert-mutations.ts
               graphql-alert-types.ts
            index.ts
        
         rest/
            device/
               rest-device-endpoints.ts
            commands/
               rest-command-endpoints.ts
            index.ts
        
         mock/
            mock-device-data.ts
            mock-command-data.ts
            mock-log-data.ts
        
         index.ts              # API exports
```

### 5.2 Import Rules Visualized

```
src/
 ui/                  ← CAN import from: hooks/
    pages/          ← CANNOT import from: domain/, lib/api/
    components/     ← CANNOT import from: domain/, lib/api/

 hooks/             ← CAN import from: domain/, lib/api/
    *.ts           ← CANNOT import from: ui/, hooks/

 domain/            ← CAN import from: NOTHING
    *.ts           ← CANNOT import from: ui/, hooks/, lib/api/

 lib/
     api/           ← CAN import from: domain/ (types only)
         *.ts       ← CANNOT import from: ui/, hooks/
```

---

## 6. Domain Layer

### 6.1 Purpose

The domain layer is the **pure core** of the application. It defines what the application is, without any implementation details.

### 6.2 What Belongs in Domain

| Type | Examples |
|------|----------|
| **Type Definitions** | `Device`, `Command`, `LogEntry` |
| **Interfaces** | Repository interfaces |
| **Pure Functions** | Data transformations, validations |
| **Constants** | Command presets, status enums |

### 6.3 What Does NOT Belong in Domain

| Excluded | Reason |
|---------|--------|
| API calls | Belongs in Data layer |
| React hooks | Belongs in Presentation layer |
| Component code | Belongs in UI layer |
| Side effects | Domain must be pure |

### 6.4 Example: Device Domain

```typescript
// packages/API_Client/src/domain/device/types.ts

export type DeviceStatus = "online" | "offline" | "deregistered";

export interface Device {
  id: string;
  imei: string;
  deviceName: string;
  model: string | null;
  manufacturer: string | null;
  osVersion: string;
  appVersion: string;
  status: DeviceStatus;
  registeredAt: Date | null;
  lastSeen: Date | null;
}

export interface DeviceListItem {
  id: string;
  imei: string;
  deviceName: string;
  status: DeviceStatus;
  lastSeen: Date | null;
}
```

```typescript
// packages/API_Client/src/domain/device/transforms.ts

import type { Device, DeviceListItem } from "./types";

interface RawDevice {
  id: string;
  imei: string;
  deviceName: string | null;
  model: string | null;
  manufacturer: string | null;
  osVersion: string;
  appVersion: string;
  online: boolean;
  registeredAt: string | null;
  lastSeen: string | null;
}

export const deviceFromRaw = (raw: RawDevice): Device => ({
  id: raw.id,
  imei: raw.imei,
  deviceName: raw.deviceName ?? "Unknown Device",
  model: raw.model,
  manufacturer: raw.manufacturer,
  osVersion: raw.osVersion,
  appVersion: raw.appVersion,
  status: raw.online ? "online" : "offline",
  registeredAt: raw.registeredAt ? new Date(raw.registeredAt) : null,
  lastSeen: raw.lastSeen ? new Date(raw.lastSeen) : null,
});

export const deviceListItemFromRaw = (raw: RawDevice): DeviceListItem => ({
  id: raw.id,
  imei: raw.imei,
  deviceName: raw.deviceName ?? "Unknown Device",
  status: raw.online ? "online" : "offline",
  lastSeen: raw.lastSeen ? new Date(raw.lastSeen) : null,
});
```

```typescript
// packages/API_Client/src/domain/device/validation.ts

export const validateImei = (imei: string): boolean => {
  // IMEI must be 15 digits
  return /^\d{15}$/.test(imei);
};

export const validateDeviceName = (name: string): boolean => {
  // Device name must be 1-100 characters
  return name.length >= 1 && name.length <= 100;
};
```

```typescript
// packages/API_Client/src/domain/device/index.ts

export * from "./types";
export * from "./transforms";
export * from "./validation";
```

---

## 7. Data Layer

### 7.1 Purpose

The data layer handles all **external communication**. It knows nothing about UI or presentation - it only knows how to make API requests and parse responses.

### 7.2 GraphQL vs REST

| Aspect | GraphQL | REST |
|--------|---------|------|
| **Primary Use** | Default for all queries/mutations | Fallback when GraphQL unavailable |
| **Client** | `@apollo/client` or `urql` | `fetch` with typed wrappers |
| **Queries** | Typed queries in `graphql/queries/` | Typed functions in `rest/endpoints.ts` |

### 7.3 GraphQL Client Setup

```typescript
// packages/API_Client/src/vyzorServer/graphql/client.ts

import { ApolloClient, InMemoryCache, HttpLink } from "@apollo/client";

export const graphqlClient = new ApolloClient({
  link: new HttpLink({
    uri: import.meta.env.VITE_GRAPHQL_URL,
    headers: {
      Authorization: `Bearer ${getAuthToken()}`,
    },
  }),
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: "cache-and-network",
    },
  },
});
```

### 7.4 GraphQL Queries

```typescript
// packages/API_Client/src/vyzorServer/graphql/queries/device.ts

import { gql } from "@apollo/client";

export const GET_DEVICES = gql`
  query GetDevices($page: Int, $limit: Int) {
    devices(page: $page, limit: $limit) {
      devices {
        id
        imei
        deviceName
        model
        manufacturer
        osVersion
        appVersion
        online
        registeredAt
        lastSeen
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
`;

export const GET_DEVICE = gql`
  query GetDevice($imei: String!) {
    device(imei: $imei) {
      id
      imei
      deviceName
      model
      manufacturer
      osVersion
      appVersion
      online
      registeredAt
      lastSeen
    }
  }
`;
```

### 7.5 REST Client

```typescript
// packages/API_Client/src/vyzorServer/rest/client.ts

const getBaseUrl = () => import.meta.env.VITE_API_URL;

const getHeaders = () => ({
  "Content-Type": "application/json",
  Authorization: `Bearer ${getAuthToken()}`,
});

export const restClient = {
  async get<T>(endpoint: string): Promise<T> {
    const response = await fetch(`${getBaseUrl()}${endpoint}`, {
      method: "GET",
      headers: getHeaders(),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  async post<T>(endpoint: string, body: unknown): Promise<T> {
    const response = await fetch(`${getBaseUrl()}${endpoint}`, {
      method: "POST",
      headers: getHeaders(),
      body: JSON.stringify(body),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    return response.json();
  },
  
  async delete<T>(endpoint: string): Promise<T> {
    const response = await fetch(`${getBaseUrl()}${endpoint}`, {
      method: "DELETE",
      headers: getHeaders(),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    return response.json();
  },
};
```

### 7.6 REST Endpoints

```typescript
// packages/API_Client/src/vyzorServer/rest/endpoints.ts

import { restClient } from "./client";

export const devicesApi = {
  list: (params?: { page?: number; limit?: number }) => {
    const query = new URLSearchParams(
      Object.entries(params ?? {}).map(([k, v]) => [k, String(v)])
    );
    return restClient.get<DeviceListResponse>(`/v1/device?${query}`);
  },
  
  get: (imei: string) => {
    return restClient.get<DeviceResponse>(`/v1/device/${imei}`);
  },
};

export const commandsApi = {
  send: (imei: string, command: string) => {
    return restClient.post<CommandResponse>(`/v1/device/${imei}/command`, {
      command,
    });
  },
  
  list: (imei: string, params?: { status?: string; page?: number }) => {
    const query = new URLSearchParams(
      Object.entries(params ?? {}).map(([k, v]) => [k, String(v)])
    );
    return restClient.get<CommandListResponse>(
      `/v1/device/${imei}/commands?${query}`
    );
  },
};
```

---

## 8. Presentation Layer

### 8.1 Purpose

The presentation layer contains all **React hooks** that:
- Transform data from the data layer for UI consumption
- Handle UI logic (loading states, errors, etc.)
- Manage component state
- NEVER render UI

### 8.2 Hook Structure

```typescript
// Every hook should follow this pattern:

export const use[Feature] = (/* params */) => {
  // 1. Get raw data from data layer
  const { data, isLoading, error } = use[Feature]Query(params);
  
  // 2. Transform raw data to domain types
  const items = data?.items.map(itemFromRaw) ?? [];
  
  // 3. Compute derived state
  const isEmpty = !isLoading && items.length === 0;
  
  // 4. Return everything UI needs
  return {
    items,
    isLoading,
    error,
    isEmpty,
    // Any other UI-friendly state
  };
};
```

### 8.3 Example: useDevices Hook

```typescript
// src/hooks/device/use-devices.ts

import { useQuery } from "@tanstack/react-query";
import { graphqlClient } from "@vyzorix/api-client/vyzorServer/graphql/client";
import { GET_DEVICES } from "@vyzorix/api-client/vyzorServer/graphql/queries/device";
import { 
  deviceFromRaw, 
  deviceListItemFromRaw,
  validateDeviceListResponse,
} from "@vyzorix/api-client/domain/device";
import type { Device, DeviceListItem } from "@vyzorix/api-client/domain/device";

interface UseDevicesOptions {
  page?: number;
  limit?: number;
  enabled?: boolean;
}

interface UseDevicesResult {
  devices: DeviceListItem[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  } | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => void;
}

export const useDevices = (
  options: UseDevicesOptions = {}
): UseDevicesResult => {
  const { page = 1, limit = 20, enabled = true } = options;
  
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ["devices", page, limit],
    queryFn: async () => {
      const response = await graphqlClient.query({
        query: GET_DEVICES,
        variables: { page, limit },
        fetchPolicy: "network-only",
      });
      
      // Validate response at boundary
      validateDeviceListResponse(response.data);
      
      return response.data.devices;
    },
    enabled,
  });
  
  // Transform at presentation layer boundary
  const devices: DeviceListItem[] = data?.devices.map(deviceListItemFromRaw) ?? [];
  
  const pagination = data?.pagination ?? null;
  
  return {
    devices,
    pagination,
    isLoading,
    error: error as Error | null,
    refetch,
  };
};
```

### 8.4 Example: useSendCommand Hook

```typescript
// src/hooks/commands/use-send-command.ts

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { graphqlClient } from "@vyzorix/api-client/vyzorServer/graphql/client";
import { SEND_COMMAND } from "@vyzorix/api-client/vyzorServer/graphql/mutations/command-mutations";
import { validateCommandResponse } from "@vyzorix/api-client/domain/commands";
import type { CommandResponse } from "@vyzorix/api-client/domain/commands";

interface UseSendCommandOptions {
  imei: string;
  onSuccess?: (response: CommandResponse) => void;
  onError?: (error: Error) => void;
}

interface SendCommandInput {
  imei: string;
  command: string;
}

export const useSendCommand = (options: UseSendCommandOptions) => {
  const { imei, onSuccess, onError } = options;
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (command: string): Promise<CommandResponse> => {
      const response = await graphqlClient.mutate({
        mutation: SEND_COMMAND,
        variables: { imei, command },
      });
      
      // Validate at boundary
      validateCommandResponse(response.data);
      
      return response.data!.sendCommand;
    },
    onSuccess: (data) => {
      // Invalidate commands cache
      queryClient.invalidateQueries({
        queryKey: ["commands", imei],
      });
      
      onSuccess?.(data);
    },
    onError: (error) => {
      onError?.(error as Error);
    },
  });
};
```

### 8.5 Example: useExport Hook

```typescript
// src/hooks/export/use-export.ts

import { useState, useCallback } from "react";
import { 
  toCSV, 
  toJSON, 
  generateFilename,
  validateExportData,
} from "@vyzorix/api-client/domain/export";
import type { ExportFormat, ExportableData } from "@vyzorix/api-client/domain/export";

interface UseExportOptions {
  type: string;
  deviceName: string;
  onExportStart?: () => void;
  onExportComplete?: () => void;
  onExportError?: (error: Error) => void;
}

export const useExport = (options: UseExportOptions) => {
  const { type, deviceName, onExportStart, onExportComplete, onExportError } = options;
  const [isExporting, setIsExporting] = useState(false);
  
  const exportData = useCallback(
    async (data: ExportableData[], format: ExportFormat) => {
      setIsExporting(true);
      onExportStart?.();
      
      try {
        // Validate data
        validateExportData(data);
        
        // Generate content
        const content = format === "csv" 
          ? toCSV(data) 
          : toJSON(data);
        
        // Generate filename
        const filename = generateFilename(type, deviceName, format);
        
        // Create download
        const blob = new Blob([content], { 
          type: format === "csv" 
            ? "text/csv;charset=utf-8" 
            : "application/json" 
        });
        
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = filename;
        link.click();
        URL.revokeObjectURL(url);
        
        onExportComplete?.();
      } catch (error) {
        onExportError?.(error as Error);
      } finally {
        setIsExporting(false);
      }
    },
    [type, deviceName, onExportStart, onExportComplete, onExportError]
  );
  
  return {
    exportData,
    isExporting,
  };
};
```

---

## 9. UI Layer

### 9.1 Purpose

The UI layer contains **only rendering code**:
- Pages
- Components
- Shared UI elements

### 9.2 Rules

1. **NEVER** import from `lib/api/` (Data layer)
2. **NEVER** import from `domain/` (Domain layer)
3. **ONLY** import from `hooks/` (Presentation layer) and other UI components
4. **ONLY** call hook functions, don't access hook internals

### 9.3 Page Example

```typescript
// src/ui/pages/commands/commands-send.tsx

import { useState } from "react";
import { useSendCommand } from "@/hooks/commands/use-send-command";
import { useCommandsPending } from "@/hooks/commands/use-commands";
import { useDeviceSelector } from "@/hooks/common/use-device-selector";
import { PRESET_COMMANDS } from "@vyzorix/api-client/domain/commands";
import { toast } from "sonner";

import { Section } from "@/ui/components/ui/section";
import { Button } from "@/ui/components/ui/button";
import { CommandButton } from "@/ui/components/shared/command-button";
import { CommandsRecent } from "@/ui/components/shared/commands-recent";
import { DeviceSelector } from "@/ui/components/shared/device-selector";

export const CommandsSendPage = () => {
  const { selectedDevice } = useDeviceSelector();
  const { pendingCommands } = useCommandsPending(selectedDevice?.imei);
  const { sendCommand, isSending } = useSendCommand({
    imei: selectedDevice?.imei ?? "",
    onSuccess: (response) => {
      toast.success("Command sent", {
        description: `${response.command} → ${response.dispatchId.slice(0, 8)}`,
      });
    },
    onError: (error) => {
      toast.error("Failed to send command", {
        description: error.message,
      });
    },
  });
  
  const handleSendCommand = (commandId: string) => {
    if (!selectedDevice) {
      toast.error("Please select a device first");
      return;
    }
    sendCommand(commandId);
  };
  
  return (
    <div className="space-y-6">
      {/* Device Selector */}
      <Section>
        <Section.Header title="Target Device" />
        <Section.Content>
          <DeviceSelector />
        </Section.Content>
      </Section>
      
      {/* Command Grid */}
      <Section>
        <Section.Header title="Send Command" />
        <Section.Content>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {PRESET_COMMANDS.map((command) => (
              <CommandButton
                key={command.id}
                command={command}
                onClick={() => handleSendCommand(command.id)}
                disabled={!selectedDevice || isSending}
                loading={isSending}
              />
            ))}
          </div>
        </Section.Content>
      </Section>
      
      {/* Pending Commands */}
      {pendingCommands.length > 0 && (
        <Section>
          <Section.Header 
            title="Pending" 
            badge={`${pendingCommands.length}`} 
          />
          <Section.Content>
            <div className="space-y-2">
              {pendingCommands.map((command) => (
                <CommandRow
                  key={command.dispatchId}
                  command={command}
                  showCancel
                />
              ))}
            </div>
          </Section.Content>
        </Section>
      )}
      
      {/* Recent Commands */}
      <CommandsRecent imei={selectedDevice?.imei} limit={5} />
    </div>
  );
};
```

### 9.4 Component Example

```typescript
// src/ui/components/shared/command-button/command-button.tsx

import { ButtonHTMLAttributes } from "react";
import { AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/ui/components/ui/button";
import type { PresetCommand } from "@vyzorix/api-client/domain/commands";

interface CommandButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  command: PresetCommand;
  loading?: boolean;
}

export const CommandButton = ({
  command,
  loading = false,
  className,
  ...props
}: CommandButtonProps) => {
  return (
    <Button
      variant={command.danger ? "destructive" : "default"}
      size="lg"
      className={cn(
        "h-auto flex-col gap-2 py-4",
        command.danger && "border-destructive/50",
        className
      )}
      disabled={loading || props.disabled}
      {...props}
    >
      {command.danger && (
        <AlertTriangle className="h-4 w-4" />
      )}
      <span className="font-medium">{command.label}</span>
      <span className="text-xs opacity-70">{command.description}</span>
    </Button>
  );
};
```

---

## 10. Implementation Guide

### 10.1 Creating a New Feature

Follow this order:

```

                                                                     
  STEP 1: Define Domain                                             
     
                                                                     
  packages/API_Client/src/domain/[feature]/                                             
   types.ts          # What is this thing?                         
   transforms.ts     # How do I convert API data?                  
   validation.ts     # Is this data valid?                         
                                                                     
  ↓ (next step)                                                     
                                                                     
  STEP 2: Create Data Layer                                         
     
                                                                     
  packages/API_Client/src/vyzorServer/graphql/                                              
   queries/[feature].ts    # GraphQL queries                      
   mutations/[feature].ts   # GraphQL mutations                    
                                                                     
  packages/API_Client/src/vyzorServer/rest/                                                
   endpoints.ts           # REST endpoints (fallback)             
   adapters.ts           # Convert REST to GraphQL format         
                                                                     
  ↓ (next step)                                                     
                                                                     
  STEP 3: Create Presentation Layer                                 
     
                                                                     
  src/hooks/[feature]/                                              
   use-[feature].ts          # Main hook                          
   use-[feature]-list.ts    # List hook                         
   use-[feature]-mutation.ts # Mutation hook                      
                                                                     
  ↓ (next step)                                                     
                                                                     
  STEP 4: Create UI Layer                                           
     
                                                                     
  src/ui/pages/[feature]/                                            
   [feature]-page.tsx        # Main page                          
                                                                     
  src/ui/components/shared/[feature]/                                
   [feature]-card.tsx        # Card component                     
   [feature]-row.tsx          # Row component                     
                                                                     

```

### 10.2 Example: Add Alert Feature

#### Step 1: Domain

```typescript
// packages/API_Client/src/domain/alerts/types.ts

export type AlertSeverity = "critical" | "warning" | "info";

export type AlertStatus = "active" | "acknowledged" | "resolved";

export interface Alert {
  id: string;
  deviceId: string;
  deviceName: string;
  type: string;
  severity: AlertSeverity;
  message: string;
  status: AlertStatus;
  createdAt: Date;
  acknowledgedAt: Date | null;
  resolvedAt: Date | null;
}
```

```typescript
// packages/API_Client/src/domain/alerts/transforms.ts

import type { Alert } from "./types";

interface RawAlert {
  id: string;
  deviceId: string;
  deviceName: string;
  type: string;
  severity: string;
  message: string;
  status: string;
  createdAt: string;
  acknowledgedAt: string | null;
  resolvedAt: string | null;
}

export const alertFromRaw = (raw: RawAlert): Alert => ({
  id: raw.id,
  deviceId: raw.deviceId,
  deviceName: raw.deviceName,
  type: raw.type,
  severity: raw.severity as AlertSeverity,
  message: raw.message,
  status: raw.status as AlertStatus,
  createdAt: new Date(raw.createdAt),
  acknowledgedAt: raw.acknowledgedAt ? new Date(raw.acknowledgedAt) : null,
  resolvedAt: raw.resolvedAt ? new Date(raw.resolvedAt) : null,
});
```

#### Step 2: Data Layer

```typescript
// packages/API_Client/src/vyzorServer/graphql/queries/alert-queries.ts

import { gql } from "@apollo/client";

export const GET_ALERTS = gql`
  query GetAlerts($status: AlertStatus, $page: Int, $limit: Int) {
    alerts(status: $status, page: $page, limit: $limit) {
      alerts {
        id
        deviceId
        deviceName
        type
        severity
        message
        status
        createdAt
        acknowledgedAt
        resolvedAt
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
`;
```

#### Step 3: Presentation Layer

```typescript
// src/hooks/alerts/use-alerts.ts

import { useQuery } from "@tanstack/react-query";
import { graphqlClient } from "@vyzorix/api-client/vyzorServer/graphql/client";
import { GET_ALERTS } from "@vyzorix/api-client/vyzorServer/graphql/queries/alert-queries";
import { alertFromRaw } from "@vyzorix/api-client/domain/alerts";
import type { Alert, AlertStatus } from "@vyzorix/api-client/domain/alerts";

interface UseAlertsOptions {
  status?: AlertStatus;
  page?: number;
  limit?: number;
}

export const useAlerts = (options: UseAlertsOptions = {}) => {
  const { status, page = 1, limit = 20 } = options;
  
  const { data, isLoading, error } = useQuery({
    queryKey: ["alerts", status, page, limit],
    queryFn: async () => {
      const response = await graphqlClient.query({
        query: GET_ALERTS,
        variables: { status, page, limit },
      });
      return response.data.alerts;
    },
  });
  
  const alerts = data?.alerts.map(alertFromRaw) ?? [];
  const pagination = data?.pagination ?? null;
  
  return {
    alerts,
    pagination,
    isLoading,
    error: error as Error | null,
  };
};
```

#### Step 4: UI Layer

```typescript
// src/ui/pages/alerts/alerts-active.tsx

import { useAlerts } from "@/hooks/alerts/use-alerts";
import { Section } from "@/ui/components/ui/section";
import { AlertRow } from "@/ui/components/shared/alert-row";
import { EmptyState } from "@/ui/components/ui/empty-state";
import { AlertTriangle } from "lucide-react";

export const AlertsActivePage = () => {
  const { alerts, isLoading } = useAlerts({ status: "active" });
  
  if (!isLoading && alerts.length === 0) {
    return (
      <EmptyState
        icon={AlertTriangle}
        title="No active alerts"
        description="All systems are operating normally"
      />
    );
  }
  
  return (
    <Section>
      <Section.Header title="Active Alerts" badge={`${alerts.length}`} />
      <Section.Content>
        <div className="space-y-4">
          {alerts.map((alert) => (
            <AlertRow key={alert.id} alert={alert} />
          ))}
        </div>
      </Section.Content>
    </Section>
  );
};
```

---

## 11. Debugging Guide

### 11.1 Debug Flow

```
When something breaks, work from outside in:


                                                                     
  1. UI NOT RENDERING?                                              
      Check hook return value                                     
          Is the hook being called correctly?                     
              Is the hook returning what you expect?               
                                                                     
  2. HOOK RETURNING WRONG DATA?                                      
      Check data transformation                                    
          Is the domain transform correct?                         
              Is the raw data what you expect?                    
                                                                         
  3. API CALL FAILING?                                               
      Check data layer                                           
          Is the GraphQL query correct?                           
              Is the REST endpoint correct?                        
                  Is the API responding?                            
                                                                     
  4. DATA NOT VALID?                                                 
      Check domain validation                                     
          Is the validation function correct?                      
              What does the API actually return?                   
                                                                     

```

### 11.2 Quick Debug Commands

```bash
# Check if hook returns data
console.log("Hook data:", data);
console.log("Hook loading:", isLoading);
console.log("Hook error:", error);

// Check raw API response
console.log("Raw response:", response);

// Check domain transform
console.log("Transformed:", transformFunction(raw));

// Check validation
console.log("Is valid:", validateFunction(data));
```

### 11.3 Layer-Specific Debugging

#### UI Layer
```typescript
// Add before return in component
console.log("Props received:", { /* props */ });
console.log("Hook data:", data);
```

#### Presentation Layer
```typescript
// Add in hook
console.log("Raw data from API:", rawData);
console.log("Transformed data:", transformedData);
```

#### Domain Layer
```typescript
// Add in transform function
console.log("Input:", input);
console.log("Output:", output);

// Add in validation
console.log("Validating:", value);
console.log("Result:", isValid);
```

#### Data Layer
```typescript
// Add in API call
console.log("Request:", { endpoint, variables });
console.log("Response:", response);
```

---

## 12. Migration Checklist

### 12.1 Existing Files to Migrate

| File | New Location | Status |
|------|--------------|--------|
| `vyzorix-api.ts` | Split into `domain/` + `lib/api/` | TODO |
| Component files | `src/ui/components/` | TODO |
| Hook files | `src/hooks/` | TODO |
| Route files | `src/ui/pages/` | TODO |

### 12.2 Migration Steps

```
STEP 1: Create Directory Structure

src/
 ui/
 hooks/
 domain/
 lib/api/

STEP 2: Create Domain Types

packages/API_Client/src/domain/[feature]/types.ts
packages/API_Client/src/domain/[feature]/transforms.ts

STEP 3: Create Data Layer

packages/API_Client/src/vyzorServer/graphql/queries/[feature].ts
packages/API_Client/src/vyzorServer/rest/endpoints.ts

STEP 4: Create Presentation Layer

src/hooks/[feature]/use-[feature].ts

STEP 5: Create UI Layer

src/ui/pages/[feature]/page.tsx

STEP 6: Update Imports

# Update old imports to new locations
# Run ESLint to find broken imports

STEP 7: Delete Old Files

# Delete files from old locations
# Keep backup until verified
```

### 12.3 Verification Checklist

- [ ] All components import from hooks only
- [ ] All hooks import from domain and lib/api only
- [ ] All domain files have no external imports
- [ ] All lib/api files only import from domain types
- [ ] ESLint passes with no errors
- [ ] All tests pass
- [ ] Manual testing complete

---

*Document Version: 1.0*  
*Status: Ready for Implementation*
