# 10A 12V → 40V Boost Converter Module

## Requirements

| Parameter | Value |
|-----------|-------|
| Input Voltage | 12V (LiPo) |
| Output Voltage | 40V (adjustable) |
| Output Current | 10A max |
| Output Power | 400W |
| Control | Programmable (analog) |
| Cooling | Heatsink (passive) |
| Size | Compact (~60mm × 80mm) |

---

## Power Analysis

```
Output: 40V × 10A = 400W
Input at 12V: 400W ÷ 12V = 33.3A
With 90% efficiency: 33.3A ÷ 0.9 = 37A input current
```

**Critical Design Points:**
- Input current = 37A (heavy!)
- Switch current = ~37A peak
- Inductor must handle 10A+ continuous
- MOSFETs need low Rds(on) for minimal heat

---

## Circuit Topology

```
VIN (12V LiPo) ──────────────────────────────────────┐
                                                   │
    ┌──────────────────────────────────────────────┤
    │                                              │
    │   ┌───────┐                                 │
    │   │  L1   │◄── (inductor charging)          │
    │   │ 22µH │                                 │
    │   └───┬───┘                                 │
    │       │                                     │
    │       │        ┌────────┐                    │
    │       └────────┤   Q1   │ (N-MOSFET)         │
    │                └───┬────┘                    │
    │                    │                          │
    │                PHASE                         │
    │                    │                          │
    │                    ├──────┬──────────────────┤
    │                    │      │                  │
    │               ┌────┴───┐  │                  │
    │               │   D1   │  │ (Schottky diode) │
    │               └───┬────┘  │                  │
    │                   │       │                  │
    │    ┌──────────┐  │       │                  │
    │    │    C1    │◄─┘       │                  │
    │    │ 100µF×3  │          │                  │
    │    └────┬─────┘          │                  │
    │         │               │                  │
    └─────────┼───────────────┼──────────────────┘
              │               │
           VOUT 40V          GND
```

---

## Component Selection

### 1. Boost Controller IC
**TI LM5122** - Wide Vin Synchronous Boost
- Vin: 3-65V
- External FETs for high current
- Programmable UVLO, OCP, soft-start

### 2. Power MOSFET (Q1)
**CSD18537NQ5A** (TI)
- Vds: 40V
- Id: 100A pulse
- Rds(on): 3.2mΩ
- Package: SON 5×6mm

### 3. Synchronous FET (Q2)
**CSD18537NQ5A** (same)
- Synchronous rectification

### 4. Inductor (L1)
**Custom 22µH**
- Isat: 15A+
- DCR: <10mΩ
- Shielded construction

### 5. Catch Diode (D1)
**MBR1040** (Schottky)
- Vr: 40V
- If: 10A
- Low forward drop

### 6. Output Capacitors (C1)
**3× 100µF 50V**
- Low ESR
- Ceramic or electrolytic

### 7. Current Sense
**0.02Ω 5W** (Rcs)
- For 10A: 200mV sense

---

## Pinout / Interface

| Pin | Function |
|-----|----------|
| VIN+ | 12V input (LiPo positive) |
| VIN- | 12V input (LiPo negative) |
| VOUT+ | 40V output positive |
| VOUT- | Output negative (GND) |
| V_ADJ | Voltage adjustment (0-5V → 20-50V) |
| I_ADJ | Current limit (0-5V → 1-10A) |
| EN | Enable (high = on) |
| AGND | Analog ground |

---

## Thermal Design

### Power Dissipation
| Component | Loss (est) |
|-----------|-------------|
| Q1 MOSFET | ~3-5W |
| Q2 MOSFET | ~1-2W |
| D1 Diode | ~2-3W |
| L1 Inductor | ~1-2W |
| **Total** | **~7-12W** |

### Heatsink Required
```
Rth_heatsink < 10°C/W
```

### Board Size Estimate
```
MOSFET area: 15mm × 20mm
Inductor: 15mm × 15mm
Capacitors: 20mm × 15mm
Terminals: 20mm × 10mm
Heatsink: 50mm × 40mm

Estimated PCB: 80mm × 60mm minimum
```

---

## Block Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                 12V → 40V BOOST CONVERTER                  │
│                       80mm × 60mm                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                  POWER INPUT                          │  │
│  │   VIN+ ───────►│ BATTERY │─────► ────► VIN         │  │
│  │   VIN- ───────►│  (LiPo) │─────► ────► GND         │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                  BOOST STAGE                           │  │
│  │                                                     │  │
│  │    ┌─────────┐         ┌─────────┐                   │  │
│  │    │   L1    │◄────────│   Q1    │                  │  │
│  │    │ 22µH   │         │ (FET)   │                  │  │
│  │    └────┬────┘         └────┬────┘                   │  │
│  │         │                    │                        │  │
│  │         │              PHASE │                        │  │
│  │         │                    │                        │  │
│  │         │              ┌────┴────┐                   │  │
│  │         └─────────────►│   Q2    │ (Sync FET)       │  │
│  │                        └────┬────┘                   │  │
│  │                             │                        │  │
│  └─────────────────────────────┼────────────────────────┘  │
│                                │                            │
│                                ▼                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                  OUTPUT STAGE                         │  │
│  │                                                     │  │
│  │                        ┌────┐                       │  │
│  │                        │ D1 │ (Schottky)             │  │
│  │                        └──┬─┘                       │  │
│  │                             │                        │  │
│  │    ┌─────┐  ┌─────┐  ┌─────┴─────┐               │  │
│  │    │ C1  │  │ C2  │  │    C3      │  3×100µF   │  │
│  │    └──┬──┘  └──┬──┘  └────┬──────┘               │  │
│  │         │        │         │                       │  │
│  │         └────────┼─────────┘                       │  │
│  │                  │                                   │  │
│  └──────────────────┼───────────────────────────────────┘  │
│                     │                                      │
│                  VOUT 40V                                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                 CONTROL STAGE                         │  │
│  │                                                     │  │
│  │    ┌──────────┐  ┌──────────┐  ┌──────────┐     │  │
│  │    │ LM5122   │  │ CURRENT  │  │ OPTO/    │     │  │
│  │    │CONTROLLER│  │  SENSE   │  │ ISOLATOR │     │  │
│  │    │          │  │  Rcs     │  │          │     │  │
│  │    └────┬─────┘  └────┬─────┘  └────┬─────┘     │  │
│  │         │             │             │            │  │
│  │         ▼             ▼             ▼            │  │
│  │    V_ADJ Pin    I_ADJ Pin      EN Pin         │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │                    HEATSINK                           │  │
│  │     (Attached to Q1, Q2, D1)                        │  │
│  └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## PCB Layout Principles

### High-Current Traces
- Width: 5-10mm for 35A+ traces
- Use copper pours, not traces
- Thermal relief for solderability

### Switching Node (PHASE)
- Keep short, minimize loop area
- Away from sensitive signals

### Sense Lines
- Kelvin connections for current sense
- Keep away from switching noise

---

## JLCPCB Components

| Component | Part Number | JLCPCB | Qty |
|-----------|-------------|--------|-----|
| Boost IC | LM5122 | Extended | 1 |
| MOSFET | CSD18537NQ5A | Extended | 2 |
| Schottky | MBR1040 | Extended | 1 |
| Inductor | Custom | - | 1 |
| Caps 100µF | C19540 | Basic | 3 |
| Caps 10µF | C14663 | Basic | 2 |
| Resistors | 0402 | Basic | 5 |
| Terminal | KF301-2P | Basic | 4 |
| Op-amp | LM358 | Basic | 1 |

---

*Last Updated: June 27, 2026*
