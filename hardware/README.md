# Vyzorix Hardware Design

This folder contains hardware design files for Vyzorix IoT devices, created using [TSCircuit](https://tscircuit.com/).

## Directory Structure

```
hardware/
 tscircuit/     # TSCircuit PCB designs
    index.tsx   # Main circuit definition
    dist/       # Build outputs (PCB, schematic, 3D)
    TSCCIRCUIT_WORKFLOW.md  # Full workflow guide
    QUICKREF.md             # Quick reference
 README.md       # This file
```

## Published Packages

### @tsci/vinnsedesigner.vyzorix-hardware

**URL:** https://tscircuit.com/vinnsedesigner/vyzorix-hardware

A production-grade IoT device module featuring:
- USB-C power delivery with decoupling capacitors
- Status LED with current limiting resistor
- Reset and Boot buttons for ESP32-style MCUs
- 1x4 pin header for external connections
- JLCPCB-ready with supplier part numbers

**Package:** `@tsci/vinnsedesigner.vyzorix-hardware@0.0.1`

## Getting Started

See [TSCCIRCUIT_WORKFLOW.md](tscircuit/TSCCIRCUIT_WORKFLOW.md) for detailed setup instructions.

Quick start:
```bash
cd hardware/tscircuit
bun install
bun x tsci build --svgs --pcb-png --3d
```

## Circuit Components

| Component | Description | Part Number |
|-----------|-------------|-------------|
| USB_C | USB-C connector | SmdUsbC |
| C1 | 100nF decoupling | C14663 |
| C2 | 10uF bulk cap | C28346 |
| R1 | 330R LED resistor | C25104 |
| LED_STATUS | Blue status LED | C433 |
| BTN_RESET | Reset button | C110153 |
| BTN_BOOT | Boot button | C110153 |
| JP1 | 1x4 pin header | 2.54mm |

## Design Specifications

- **Board Size:** 50mm x 40mm
- **Power:** 5V USB-C (VBUS)
- **Nets:** VBUS, GND
- **Footprint:** 0603 (passives)
- **Manufacturer:** JLCPCB compatible
