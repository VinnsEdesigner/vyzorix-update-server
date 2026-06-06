# Vyzorix Update Server - Documentation

> **Note:** This is the server repository documentation. For Android-side docs, see [VyzorixAudioRouter](https://github.com/VinnsEdesigner/VyzorixAudioRouter).

---

## Quick Links

| Document | Description |
|----------|-------------|
| [README.md](../README.md) | Main project README |
| [REPO_TREE.md](./REPO_TREE.md) | Complete repository structure |
| [UPDATE_SERVER_ARCHITECTURE_SPEC.md](./UPDATE_SERVER_ARCHITECTURE_SPEC.md) | Server architecture deep-dive |
| [Todo.md](../Todo.md) | Task tracking and priorities |

---

## Documentation Categories

### Core Architecture

- [`SYSTEM_MAP.md`](./SYSTEM_MAP.md) - System overview and component interactions
- [`UPDATE_SERVER.md`](./UPDATE_SERVER.md) - Server API endpoints
- [`UPDATE_SERVER_ARCHITECTURE_SPEC.md`](./UPDATE_SERVER_ARCHITECTURE_SPEC.md) - Internal Go server architecture
- [`DEVICE_REGISTRATION.md`](./DEVICE_REGISTRATION.md) - Device lifecycle and registration

### Backend Documentation

- [`BACKEND_BUG_FIXES.md`](./BACKEND_BUG_FIXES.md) - Bug fixes #2-15 implemented
- [`FRONTEND_BUG_FIXES.md`](./FRONTEND_BUG_FIXES.md) - Frontend bug fixes FE-1 to FE-4
- [`COMMAND_SECURITY.md`](./COMMAND_SECURITY.md) - HMAC command signing
- [`FEATURES.md`](./FEATURES.md) - Feature list

### Build & Deployment

- [`BUILD_ORDER.md`](./BUILD_ORDER.md) - Build sequence
- [`CI_CD_WORKFLOWS.md`](./CI_CD_WORKFLOWS.md) - CI/CD documentation
- [render.yaml](../render.yaml) - Render deployment blueprint
- [.env.example](../.env.example) - Environment configuration

### Legacy Reference (Android Side)

These documents are from the Android repository and kept here for reference only:

- `DOC_1_BOOTSTRAP_AND_ORCHESTRATION.md`
- `DOC_2_ACCESSIBILITY_AND_AUTOMATION_GOVERNANCE.md`
- `DOC_3_AUDIO_PIPELINE_AND_VOIP_EXEMPTIONS.md`
- `DOC_4_RESILIENCE_FALLBACKS_AND_RECOVERY.md`
- `DOC_5_DIAGNOSTICS_CRASH_FORENSICS_AND_STORAGE.md`
- `DOC_6_MEMORY_PERFORMANCE_AND_HARDWARE_MONITORING.md`
- `DOC_7_DATA_SECURITY_AND_PERSISTENCE.md`
- `DOC_8_REALTIME_C2_COMMUNICATION_AND_UPDATES.md`

### Reference

- [`GLOSSARY.md`](./GLOSSARY.md) - Terminology
- [`NAMING_RENAMES.md`](./NAMING_RENAMES.md) - Naming conventions
- [`MEDIA_PROJECTION_FLOW.md`](./MEDIA_PROJECTION_FLOW.md) - Capture pipeline
- [`NOKIA_C22_NOTES.md`](./NOKIA_C22_NOTES.md) - Device quirks
- [`SOFT_REBOOT_ANALYSIS.md`](./SOFT_REBOOT_ANALYSIS.md) - Failure analysis
- [adr/](./adr/) - Architecture Decision Records

---

## Cleanup Notes (June 2026)

Removed obsolete files:
- `VyzorixUpdate_RepoTree.md` - superseded by `REPO_TREE.md`
- `VyzorixAudioRouter_RepoTree.md` - belongs to Android repo
- `DOC_8_REALTIME_C2_COMMUNICATION_AND_UPDATES_UPDATED.md` - duplicate
- `FEATURES_UPDATED.md` - duplicate