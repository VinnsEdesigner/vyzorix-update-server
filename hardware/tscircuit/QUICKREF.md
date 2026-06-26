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

# Build with all outputs (PRODUCTION)
bun x tsci build --svgs --pcb-png --3d --kicad-project

# Verification checks
bun x tsci check netlist
bun x tsci check pin_specification

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
- dip_4, dip_8 (through-hole)
- "chip_name" (ICs - check registry)

## ⚠️ PRODUCTION RULES (Critical!)

### Component Spacing
| Component | Size | Min Separation |
|-----------|------|---------------|
| 6x6mm pushbutton | 6mm | **>10mm** center-to-center |
| 0603 SMD | 1.5x0.8mm | 0.5mm |
| USB-C | 10x4mm | 2mm |

### Board Boundaries
- Components must be **>2mm inside** board edges
- 6x6mm button center: keep >5mm from edge

### Build Success Criteria
```
✓ 0 errors (errors block build)
✓ 0 warnings about overlaps
✓ Components within board
✓ Clearance > 0.1mm
```

## Full Workflow
1. `bun x tsci clone user/project` OR manual setup
2. Edit `index.tsx` with circuit
3. `bun x tsci build --svgs --pcb-png --3d --kicad-project`
4. Run `bun x tsci check netlist` - expect 0 errors
5. Fix any overlap/spacing warnings
6. `bun x tsci push --include-dist`

## Common Errors → Fixes
| Error | Fix |
|--------|-----|
| `pcb_plated_hole overlaps` | Move components >10mm apart |
| `Courtyard overlaps` | Increase spacing |
| `extends outside board` | Move 2mm inward |
| `footprint does not match` | IoU<0.5 fix for production |
