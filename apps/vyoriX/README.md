# vyoriX - Frontend Blueprint

Complete directory structure as specified in `documents/FRONTEND_ARCHITECTURE.md`.

## Structure

```
src/
├── ui/                              # UI LAYER
│   ├── pages/                       # Route pages
│   │   ├── dashboard/
│   │   ├── commands/
│   │   ├── logs/
│   │   ├── device/
│   │   ├── diagnostics/
│   │   ├── alerts/
│   │   └── updates/
│   │
│   └── components/
│       ├── ui/                      # Base shadcn/ui components
│       ├── layout/                  # Layout components
│       └── shared/                  # Shared feature components
│           ├── section/
│           ├── metric-card/
│           ├── connection-status/
│           ├── device-selector/
│           ├── command-button/
│           ├── command-row/
│           ├── log-entry/
│           └── export-menu/
│
├── hooks/                           # PRESENTATION LAYER
│   ├── auth/
│   ├── device/
│   ├── commands/
│   ├── logs/
│   ├── alerts/
│   ├── telemetry/
│   ├── export/
│   └── _shared/
│
├── domain/                          # DOMAIN LAYER
│   ├── _shared/
│   ├── device/
│   ├── commands/
│   ├── logs/
│   ├── telemetry/
│   ├── alerts/
│   └── export/
│
└── lib/
    └── api/                         # DATA LAYER
        ├── _shared/
        ├── graphql/
        │   ├── device/
        │   ├── commands/
        │   ├── logs/
        │   ├── telemetry/
        │   └── alerts/
        ├── rest/
        │   ├── device/
        │   └── commands/
        └── mock/
```

## Dependency Flow

```
UI → Hooks → Domain → API
       ↓
     (Data)
```

UI can ONLY import from Hooks.
Hooks can import from Domain and Data.
Domain can NOT import from anything.
Data can import from Domain (types only).
