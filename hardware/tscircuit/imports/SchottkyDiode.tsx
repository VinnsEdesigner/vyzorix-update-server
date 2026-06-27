/**
 * SchottkyDiode - Schottky Barrier Diode
 *
 * 1A Schottky diode with forward voltage ~0.3V
 * Used for bootstrap and free-wheeling applications
 *
 * Pins:
 * - pin1: Anode (positive)
 * - pin2: Cathode (negative, marked with bar)
 *
 * @category Diodes
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const diodePinLabels = {
  pin1: ["A", "ANODE"],
  pin2: ["K", "CATHODE", "KATH"]
} as const

export const SchottkyDiode = (props: ChipProps<typeof diodePinLabels>) => (
  <chip
    pinLabels={diodePinLabels}
    pinAttributes={{
      A: { requiresPower: true },
      K: { requiresPower: true },
    }}
    manufacturerPartNumber="MBRS140"
    supplierPartNumbers={{ jlcpcb: ["C96789"] }}
    footprint={
      <footprint>
        {/* Anode pad (left) */}
        <smtpad portHints={["pin1"]} pcbX="-2mm" pcbY="0mm" width="1.2mm" height="0.8mm" shape="rect" />
        {/* Cathode pad (right) */}
        <smtpad portHints={["pin2"]} pcbX="2mm" pcbY="0mm" width="1.2mm" height="0.8mm" shape="rect" />
        {/* Diode body outline */}
        <silkscreenpath route={[{ x: -3, y: -1.5 }, { x: 3, y: -1.5 }, { x: 3, y: 1.5 }, { x: -3, y: 1.5 }, { x: -3, y: -1.5 }]} />
        {/* Diode symbol */}
        <silkscreenpath route={[{ x: -1.5, y: -1 }, { x: -1.5, y: 1 }]} />
        <silkscreenpath route={[{ x: -1.8, y: -1 }, { x: -1.5, y: 0 }, { x: -1.8, y: 1 }]} />
        <silkscreenpath route={[{ x: 1.5, y: -1 }, { x: 1.5, y: 1 }]} />
        <silkscreenpath route={[{ x: 1.5, y: -1 }, { x: 1.8, y: 0 }]} />
        <silkscreenpath route={[{ x: 1.5, y: 1 }, { x: 1.8, y: 0 }]} />
        {/* Cathode bar */}
        <silkscreenpath route={[{ x: 3, y: -1.5 }, { x: 3, y: 1.5 }]} strokeWidth="0.3mm" />
        {/* Reference */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-3mm" fontSize="0.6mm" anchorAlignment="center" />
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -3.5, y: -2 },
          { x: 3.5, y: -2 },
          { x: 3.5, y: 2 },
          { x: -3.5, y: 2 },
          { x: -3.5, y: -2 }
        ]} />
      </footprint>
    }
    {...props}
  />
)