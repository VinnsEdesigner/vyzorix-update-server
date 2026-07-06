# vyoriX - Frontend Structure Blueprint

Complete directory structure based on all frontend spec documents:
- `documents/FRONTEND_ARCHITECTURE.md` (base)
- `documents/FRONTEND_API_KEYS_REQUIREMENTS.md`
- `documents/SETTINGS_PAGE.md`
- `documents/UPDATES_PAGE.md`
- `documents/DIAGNOSTICS_PAGE.md`
- `documents/DASHBOARD_COMMANDS_LOGS.md`
- `documents/DEVICE_REGISTRATION_SYSTEM.md`
- `documents/REALTIME_WEBSOCKET_ARCHITECTURE.md`

## Full Structure (77 directories)

```
src/
├── ui/                                    # UI LAYER
│   ├── pages/
│   │   ├── dashboard/
│   │   ├── commands/
│   │   ├── logs/
│   │   ├── device/
│   │   ├── diagnostics/
│   │   ├── alerts/
│   │   └── updates/
│   │
│   └── components/
│       ├── ui/                           # Base shadcn/ui
│       ├── layout/                       # Layout components
│       └── shared/                       # Cross-feature components
│           ├── section/
│           ├── metric-card/
│           ├── connection-status/
│           ├── device-selector/
│           ├── command-button/
│           ├── command-row/
│           ├── log-entry/
│           └── export-menu/
│
├── hooks/                                 # PRESENTATION LAYER
│   ├── auth/
│   ├── device/
│   ├── commands/
│   ├── logs/
│   ├── alerts/
│   ├── telemetry/
│   ├── export/
│   ├── apikey/                            # From FRONTEND_API_KEYS_REQUIREMENTS.md
│   ├── settings/                          # From SETTINGS_PAGE.md
│   ├── updates/                           # From UPDATES_PAGE.md
│   ├── diagnostics/                       # From DIAGNOSTICS_PAGE.md
│   ├── registration/                      # From DEVICE_REGISTRATION_SYSTEM.md
│   ├── realtime/                          # From REALTIME_WEBSOCKET_ARCHITECTURE.md
│   └── _shared/
│
├── domain/                                # DOMAIN LAYER
│   ├── _shared/
│   ├── device/
│   ├── commands/
│   ├── logs/
│   ├── telemetry/
│   ├── alerts/
│   ├── export/
│   ├── apikey/                            # From FRONTEND_API_KEYS_REQUIREMENTS.md
│   ├── settings/                          # From SETTINGS_PAGE.md
│   ├── updates/                           # From UPDATES_PAGE.md
│   ├── diagnostics/                       # From DIAGNOSTICS_PAGE.md
│   ├── registration/                      # From DEVICE_REGISTRATION_SYSTEM.md
│   └── realtime/                         # From REALTIME_WEBSOCKET_ARCHITECTURE.md
│
└── lib/
    └── api/                              # DATA LAYER
        ├── _shared/
        ├── graphql/
        │   ├── device/
        │   ├── commands/
        │   ├── logs/
        │   ├── telemetry/
        │   ├── alerts/
        │   ├── apikey/
        │   ├── settings/
        │   ├── updates/
        │   ├── diagnostics/
        │   ├── registration/
        │   └── realtime/
        ├── rest/
        │   ├── device/
        │   ├── commands/
        │   ├── apikey/
        │   ├── settings/
        │   ├── updates/
        │   ├── diagnostics/
        │   ├── registration/
        │   └── realtime/
        ├── websocket/
        │   └── realtime/                 # WebSocket only for realtime
        └── mock/
```

## Dependency Flow (STRICT)

```
UI → Hooks → Domain → API
```

- **UI** can ONLY import from Hooks
- **Hooks** can import from Domain and API
- **Domain** can NOT import from anything
- **API** can import from Domain (types only)

## Spec Coverage

| Spec Document | Features |
|--------------|----------|
| FRONTEND_ARCHITECTURE.md | Base structure + device, commands, logs, telemetry, alerts, export |
| FRONTEND_API_KEYS_REQUIREMENTS.md | apikey domain/hooks/api |
| SETTINGS_PAGE.md | settings domain/hooks/api |
| UPDATES_PAGE.md | updates domain/hooks/api |
| DIAGNOSTICS_PAGE.md | diagnostics domain/hooks/api |
| DASHBOARD_COMMANDS_LOGS.md | commands, logs features |
| DEVICE_REGISTRATION_SYSTEM.md | registration domain/hooks/api |
| REALTIME_WEBSOCKET_ARCHITECTURE.md | realtime domain/hooks/api + websocket |

## Implementation Order

1. **Shared first**: `domain/_shared/`, `lib/api/_shared/`, `hooks/_shared/`
2. **Device domain**: Used by almost all features
3. **Per feature**: Implement one feature at a time

See `documents/CROSS_SPEC_FILE_MAPPING.md` for exact file listings.
