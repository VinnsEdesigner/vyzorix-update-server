# TSCircuit Complete Workflow Guide

> **Reference Date:** June 26, 2026  
> **Author:** OpenHands AI Assistant  
> **Purpose:** Butter-smooth TSCircuit workflow from setup to production PCB design

---

## Table of Contents

1. [Environment Setup](#1-environment-setup)
2. [Authentication & Registry](#2-authentication--registry)
3. [Project Initialization](#3-project-initialization)
4. [Component Design](#4-component-design)
5. [Dependencies & Imports](#5-dependencies--imports)
6. [Building & Debugging](#6-building--debugging)
7. [Publishing to Registry](#7-publishing-to-registry)
8. [Troubleshooting](#8-troubleshooting)
9. [Best Practices](#9-best-practices)

---

## 1. Environment Setup

### 1.1 Install Bun Runtime

Bun is required for TSCircuit CLI operations.

```bash
# Download bun binary
curl -fsSL https://github.com/oven-sh/bun/releases/download/bun-v1.3.14/bun-linux-x64.zip -o /tmp/bun.zip

# Extract using Python (if unzip not available)
python3 -c "import zipfile; zipfile.ZipFile('/tmp/bun.zip').extractall('/tmp/bun')"

# Make executable
chmod +x /tmp/bun/bun-linux-x64/bun
sudo ln -s /tmp/bun/bun-linux-x64/bun /usr/local/bin/bun

# Verify
bun --version
```

### 1.2 Clone TSCircuit CLI Repository

```bash
git clone https://github.com/tscircuit/tscircuit-cli.git
cd tscircuit-cli
```

### 1.3 Install Dependencies

```bash
bun install
```

### 1.4 Build the CLI

```bash
bun run build
```

---

## 2. Authentication & Registry

### 2.1 Login to TSCircuit

```bash
bun run index.js login
```

This opens a browser for GitHub OAuth authentication. After login, credentials are stored in `~/.config/tscircuit/` (Linux) or equivalent.

### 2.2 Verify Authentication

```bash
bun run index.js whoami
```

### 2.3 Registry URL

- **Web Interface:** https://tscircuit.com
- **Package URL Format:** https://tscircuit.com/{username}/{package-name}
- **Package Name Format:** `@tsci/{username}.{package-name}`

---

## 3. Project Initialization

### 3.1 Method A: Clone an Existing Package

```bash
bun x tsci clone {author}/{package-name}
cd {package-name}
```

Example:
```bash
bun x tsci clone seveibar/usb-c-flashlight
cd usb-c-flashlight
```

### 3.2 Method B: Manual Setup

```bash
mkdir my-project
cd my-project
bun init -y
bun add tscircuit
```

### 3.3 Required File Structure

```
my-project/
├── index.tsx          # Main circuit file (MUST be .tsx!)
├── package.json       # Project metadata
├── tsconfig.json      # TypeScript config
└── dist/              # Build output (generated)
```

### 3.4 Essential package.json

```json
{
  "name": "@tsci/{username}.{project-name}",
  "version": "0.0.1",
  "description": "Project description",
  "module": "index.tsx",
  "type": "module",
  "scripts": {
    "build": "tsci build",
    "dev": "tsci dev",
    "push": "tsci push"
  },
  "devDependencies": {
    "@types/bun": "latest",
    "@types/react": "^19.2.17",
    "tscircuit": "^0.0.1963"
  },
  "keywords": ["tscircuit", "pcb", "hardware"],
  "author": "@{username}",
  "license": "MIT"
}
```

### 3.5 TypeScript Configuration

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "jsx": "react-jsx",
    "outDir": "dist",
    "strict": true,
    "esModuleInterop": true,
    "moduleResolution": "node",
    "skipLibCheck": true,
    "types": ["tscircuit"]
  },
  "include": ["**/*.ts", "**/*.tsx"]
}
```

---

## 4. Component Design

### 4.1 File Naming

⚠️ **CRITICAL:** Circuit files MUST use `.tsx` extension, not `.ts`.

```bash
# WRONG
mv index.ts index.tsx

# RIGHT
# Always create with .tsx extension
```

### 4.2 Basic Circuit Structure

```tsx
import { ComponentName } from "@tsci/{author}.{package}"

export default () => {
  return (
    <board width="50mm" height="40mm">
      {/* Components go here */}
    </board>
  )
}
```

### 4.3 Component Placement

#### Built-in Components

| Component | Usage | Props |
|-----------|-------|-------|
| `<board>` | Root element | `width`, `height`, `backgroundColor` |
| `<net>` | Define power nets | `name` |
| `<resistor>` | Resistor | `name`, `footprint`, `resistance` |
| `<capacitor>` | Capacitor | `name`, `footprint`, `capacitance` |
| `<led>` | LED | `name`, `footprint`, `ledColor` |
| `<chip>` | Generic IC | `name`, `footprint`, `manufacturerPartNumber` |
| `<trace>` | Wire connections | `from`, `to` |

#### Third-Party Components

```tsx
import { PushButton } from "@tsci/seveibar.push-button"
import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"
```

### 4.4 Common Component Props

| Prop | Type | Description |
|------|------|-------------|
| `name` | string | Unique identifier |
| `pcbX`, `pcbY` | number | PCB position in mm |
| `schX`, `schY` | number | Schematic position |
| `footprint` | string | Package type (0603, 0805, etc.) |
| `connections` | object | Pin-to-net mapping |
| `supplierPartNumbers` | object | LCSC/JLCPCB part numbers |

### 4.5 Net Naming Rules

⚠️ **IMPORTANT:**
- Net names CANNOT start with numbers (e.g., `3V3` is invalid)
- Net names CANNOT contain `+` or `-`
- Use underscores instead: `V3V3`, `V_BUS`, `VBUS`

```tsx
// WRONG
<net name="3V3" />
<net name="VCC-5V" />

// CORRECT
<net name="V3V3" />
<net name="VCC_5V" />
<net name="VBUS" />
```

### 4.6 Pin Selection Syntax

```tsx
// Component pin by name
".R1 .pos"

// Component pin by number
".C1 .pin1"
".LED_STATUS .neg"

// Direct net reference
"net.VBUS"
"net.GND"

// All aliases work:
".R1 .pos" == ".R1 .pin1" == ".R1 .anode" == ".R1 .left" == ".R1 .1"
".R1 .neg" == ".R1 .pin2" == ".R1 .cathode" == ".R1 .right" == ".R1 .2"
```

### 4.7 Example: Complete IoT Board

```tsx
import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"
import { PushButton } from "@tsci/seveibar.push-button"

/**
 * Vyzorix IoT Device Module
 * 
 * Features:
 * - USB-C power delivery
 * - Status LED with current limiting
 * - Reset and Boot buttons
 * - Power decoupling capacitors
 * 
 * @author @username
 */
export default () => {
  return (
    <board width="50mm" height="40mm">
      
      {/* Power Nets */}
      <net name="VBUS" />
      <net name="GND" />

      {/* USB-C Connector */}
      <SmdUsbC
        name="USB_C"
        connections={{
          GND1: "net.GND",
          GND2: "net.GND",
          VBUS1: "net.VBUS",
          VBUS2: "net.VBUS",
        }}
        pcbX={-18}
        pcbY={8}
        schX={-15}
        schY={0}
      />

      {/* Decoupling Capacitors */}
      <capacitor
        name="C1"
        footprint="0603"
        capacitance="100nF"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={-8}
        pcbY={5}
        schX={-5}
        schY={5}
      />
      <trace from=".C1 .pos" to="net.VBUS" />
      <trace from=".C1 .neg" to="net.GND" />

      {/* Status LED Circuit */}
      <resistor
        name="R1"
        footprint="0603"
        resistance="330R"
        supplierPartNumbers={{ jlcpcb: ["C25104"] }}
        pcbX={8}
        pcbY={8}
        schX={10}
        schY={5}
      />
      <led
        name="LED_STATUS"
        footprint="0603"
        ledColor="blue"
        supplierPartNumbers={{ jlcpcb: ["C433"] }}
        pcbX={15}
        pcbY={8}
        schX={15}
        schY={5}
      />
      <trace from=".R1 .pin1" to="net.VBUS" />
      <trace from=".R1 .pin2" to=".LED_STATUS .pos" />
      <trace from=".LED_STATUS .neg" to="net.GND" />

      {/* Push Buttons */}
      <PushButton
        name="BTN_RESET"
        footprint="pushbutton"
        pcbX={-5}
        pcbY={-10}
        schX={-5}
        schY={-10}
      />
      <trace from=".BTN_RESET .pin1" to="net.GND" />
      <trace from=".BTN_RESET .pin2" to="net.GND" />
    </board>
  )
}
```

---

## 5. Dependencies & Imports

### 5.1 Adding Third-Party Components

```bash
# Using tsci add (recommended)
bun x tsci add {author}/{package-name}

# Examples
bun x tsci add seveibar/smd-usb-c
bun x tsci add seveibar/push-button
bun x tsci add seveibar/esp32-module
```

### 5.2 Search Available Components

Use the MCP tool or browse the registry:
- **Registry:** https://tscircuit.com
- **Search:** https://tscircuit.com/search

### 5.3 Common Component Packages

| Package | Components | Author |
|---------|-----------|--------|
| `@tsci/seveibar.smd-usb-c` | SmdUsbC | seveibar |
| `@tsci/seveibar.push-button` | PushButton | seveibar |
| `@tsci/seveibar.usb-c-flashlight` | Full circuit example | seveibar |

### 5.4 Import Patterns

```tsx
// Single import
import { ComponentName } from "@tsci/author.package"

// Multiple imports
import { ComponentA } from "@tsci/author.package-a"
import { ComponentB } from "@tsci/author.package-b"

// Renamed imports
import { PushButton as MyButton } from "@tsci/seveibar.push-button"
```

---

## 6. Building & Debugging

### 6.1 Basic Build

```bash
bun x tsci build
```

### 6.2 Build with Outputs

```bash
# All outputs
bun x tsci build --svgs --pcb-png --3d --kicad-project

# Specific outputs
bun x tsci build --svgs           # SVG files
bun x tsci build --pcb-png        # PCB image
bun x tsci build --3d             # 3D preview
bun x tsci build --kicad-project  # KiCad files
```

### 6.3 Build Output Location

```
dist/
├── index/
│   ├── circuit.json        # Circuit data
│   ├── pcb.svg             # PCB schematic
│   ├── pcb.png             # PCB image
│   ├── schematic.svg       # Schematic diagram
│   ├── 3d.png              # 3D preview
│   └── kicad/              # KiCad project files
├── circuit.json            # Main circuit JSON
├── pcb.svg
├── schematic.svg
└── package.tgz             # For publishing
```

### 6.4 Understanding Build Warnings

These warnings don't fail the build but indicate areas for improvement:

```bash
# Pin attribute warnings
All pins on USB_C are underspecified (no pinAttributes set)
  → Solution: Add pinLabels with pinAttributes

# Footprint matching warnings
footprint "0603" does not match supplier footprint jlcpcb:C25104
  → Solution: Use exact footprint from supplier library
  → Or: Ignore for prototyping, fix for production

# Trace warnings
Port pin1 on R1 is missing a trace
  → Solution: Add missing trace connections
```

### 6.5 Common Build Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `Cannot find module '@tsci/...'` | Package not installed | `bun x tsci add author/package` |
| `Expected ">" but found "width"` | Wrong file extension | Rename `.ts` to `.tsx` |
| `Net name "3V3" cannot start with a number` | Invalid net name | Use `V3V3` instead |
| `Invalid footprint prop` | Invalid footprint string | Check valid footprints |
| `Cannot find port for selector` | Wrong pin name | Check available pins |

### 6.6 Developer Mode

```bash
bun x tsci dev
```

Opens interactive preview with hot reload.

---

## 7. Publishing to Registry

### 7.1 Update package.json

```json
{
  "name": "@tsci/{username}.{project-name}",
  "version": "0.0.1",
  "description": "Clear description of what this does",
  "keywords": ["tscircuit", "pcb", "iot", "esp32"],
  "author": "@{username}",
  "license": "MIT"
}
```

### 7.2 Create README.md

```markdown
# Project Name

Brief description of your PCB project.

## Features

- Feature 1
- Feature 2

## Usage

\`\`\`
bun install
bun run build
\`\`\`

## Hardware

- MCU: ESP32-S3
- Power: USB-C 5V
- Dimensions: 50mm x 40mm

## License

MIT
```

### 7.3 Push to Registry

```bash
# Create new package
bun x tsci push --include-dist
# When prompted, type 'y' to create new package

# Update existing package
bun x tsci push --include-dist
```

### 7.4 Push Options

```bash
bun x tsci push [file] [options]

Options:
  --private            Make the package private
  --version-tag <tag>  Publish as non-latest version
  --include-dist       Include dist directory
  --compress           Compress for faster upload
```

### 7.5 Post-Publish

- Package URL: `https://tscircuit.com/{username}/{project-name}`
- Available actions: Star, Fork, Order (if PCB service available)

---

## 8. Troubleshooting

### 8.1 Module Resolution Issues

```bash
# Clear and reinstall
rm -rf node_modules bun.lockb
bun install

# Add missing dependencies
bun add tscircuit
bun add @tsci/seveibar.push-button
```

### 8.2 Bun Installation Issues

If `unzip` is not available:
```bash
# Use Python to extract
python3 -c "import zipfile; zipfile.ZipFile('bun.zip').extractall('/tmp/bun')"
```

### 8.3 Authentication Issues

```bash
# Re-authenticate
rm -rf ~/.config/tscircuit/
bun run index.js login
```

### 8.4 Build Fails with Errors

1. Check file extension is `.tsx` not `.ts`
2. Verify all imports are correct
3. Ensure net names don't start with numbers
4. Check component prop names are valid

### 8.5 3D Model Generation Fails

Some custom footprints don't have 3D models:
- Warnings are non-blocking
- PCB and schematic SVGs will still generate
- For production, use standard footprints

---

## 9. Best Practices

### 9.1 Component Organization

```tsx
// Group by function
export default () => (
  <board>
    {/* Power Section */}
    <SmdUsbC ... />
    <capacitor ... /> {/* decoupling */}
    
    {/* Controller Section */}
    <chip ... /> {/* MCU */}
    <capacitor ... /> {/* bypass */}
    
    {/* I/O Section */}
    <led ... />
    <pushbutton ... />
  </board>
)
```

### 9.2 Net Naming Convention

```
VBUS     - Main power rail (5V from USB)
V3V3     - 3.3V regulated rail
GND      - Ground reference
NET_SCL  - I2C clock
NET_SDA  - I2C data
```

### 9.3 Component Positioning

- Place related components near each other
- Leave room for routing
- Consider PCB size constraints
- Use schX/schY for schematic layout
- Use pcbX/pcbY for PCB layout

### 9.4 Supplier Part Numbers

Always include JLCPCB part numbers for production:
```tsx
supplierPartNumbers={{ jlcpcb: ["C14663"] }}
```

### 9.5 Documentation

```tsx
/**
 * Component Name
 * 
 * Detailed description of purpose and operation.
 * 
 * @param propName - Description of prop
 * @author @{username}
 */
```

### 9.6 Version Control

```bash
# Initialize git
git init
git add .
git commit -m "Initial commit: IoT module v0.0.1"

# Tag releases
git tag v0.0.1
git push origin main --tags
```

---

## Quick Reference Card

| Task | Command |
|------|---------|
| Login | `bun run index.js login` |
| Clone | `bun x tsci clone author/package` |
| Add component | `bun x tsci add author/package` |
| Build | `bun x tsci build --svgs --pcb-png --3d` |
| Dev server | `bun x tsci dev` |
| Push | `bun x tsci push --include-dist` |
| Help | `bun x tsci --help` |

---

## Error Code Reference

| Code | Message | Solution |
|------|---------|----------|
| E001 | File must be .tsx | Rename file |
| E002 | Module not found | Run `bun x tsci add` |
| E003 | Invalid net name | Use alphanumeric + underscore |
| E004 | Missing trace | Add trace connections |
| E005 | Auth failed | Re-run `bun run index.js login` |
| E006 | Package exists | Use different name or update version |

---

*Document Version: 1.0*  
*Last Updated: June 26, 2026*
