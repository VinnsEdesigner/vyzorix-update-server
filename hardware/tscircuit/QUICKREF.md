# TSCircuit Quick Reference

## Essential Commands
```bash
# Setup bun
curl -fsSL https://bun.sh/install | bash

# Clone tscircuit-cli
git clone https://github.com/tscircuit/tscircuit-cli.git && cd tscircuit-cli && bun install

# Login
bun run index.js login

# Clone a project
bun x tsci clone seveibar/usb-c-flashlight

# Add component
bun x tsci add seveibar/push-button

# Build with all outputs
bun x tsci build --svgs --pcb-png --3d --kicad-project

# Push to registry
bun x tsci push --include-dist
```

## File Extension Rule
⚠️ **ALWAYS use `.tsx` extension, never `.ts`**

## Net Naming Rules
- ❌ `3V3`, `VCC-5V`, `Net+`
- ✅ `V3V3`, `VCC_5V`, `VBUS`

## Pin Selection
```tsx
".R1 .pin1"    // By number
".R1 .pos"     // By alias (pos = pin1)
".R1 .anode"   // By alias
```

## Import Pattern
```tsx
import { ComponentName } from "@tsci/author.package-name"
```

## Standard Footprints
- 0402, 0603, 0805 (passives)
- pushbutton (switches)
- "chip_name" (ICs - check registry)

## Full Workflow
1. `bun x tsci clone user/project` OR manual setup
2. Edit `index.tsx` with circuit
3. `bun x tsci build --svgs --pcb-png --3d`
4. Fix warnings/errors
5. `bun x tsci push --include-dist`
