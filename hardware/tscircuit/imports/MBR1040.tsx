/**
 * MBR1040 - 40V 10A Schottky Barrier Rectifier
 * 
 * Low forward drop Schottky diode for boost converter output stage
 * Vr: 40V, If: 10A, Low forward voltage drop
 * Package: SMC (DO-214AB)
 * 
 * Pins:
 * - A: Anode (positive)
 * - K: Cathode (negative)
 * 
 * @category Power
 * @package MBR1040
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const schottkyPinLabels = {
  pin1: ["A", "ANODE"],
  pin2: ["K", "CATHODE"]
} as const

export const MBR1040 = (props: ChipProps<typeof schottkyPinLabels>) => (
  <chip
    pinLabels={schottkyPinLabels}
    pinAttributes={{
      A: { requiresPower: true },
      K: { requiresPower: true },
    }}
    manufacturerPartNumber="MBR1040"
    supplierPartNumbers={{ jlcpcb: ["C11377"] }}
    footprint={
      <footprint>
        {/* Anode pad (left) */}
        <smtpad portHints={["pin1"]} pcbX="-2mm" pcbY="0mm" width="1.8mm" height="3mm" shape="rect" />
        
        {/* Cathode pad (right) */}
        <smtpad portHints={["pin2"]} pcbX="2mm" pcbY="0mm" width="1.8mm" height="3mm" shape="rect" />
        
        {/* SMC/DO-214AB outline */}
        <silkscreenpath route={[{ x: -3.5, y: -1.8 }, { x: 3.5, y: -1.8 }, { x: 3.5, y: 1.8 }, { x: -3.5, y: 1.8 }, { x: -3.5, y: -1.8 }]} />
        
        {/* Diode symbol - anode side */}
        <silkscreenpath route={[{ x: -1.5, y: -0.8 }, { x: -1.5, y: 0.8 }]} />
        <silkscreenpath route={[{ x: -1.8, y: -0.8 }, { x: -1.5, y: 0 }, { x: -1.8, y: 0.8 }]} />
        
        {/* Diode symbol - cathode side */}
        <silkscreenpath route={[{ x: 1.5, y: -0.8 }, { x: 1.5, y: 0.8 }]} />
        <silkscreenpath route={[{ x: 1.5, y: -0.8 }, { x: 1.8, y: 0 }]} />
        <silkscreenpath route={[{ x: 1.5, y: 0.8 }, { x: 1.8, y: 0 }]} />
        
        {/* Cathode bar (flat side marking) */}
        <silkscreenpath route={[{ x: 3.5, y: -1.8 }, { x: 3.5, y: 1.8 }]} strokeWidth="0.4mm" />
        
        {/* MBR1040 label */}
        <silkscreentext text="MBR1040" pcbX="0mm" pcbY="3mm" fontSize="0.5mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-3.5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -4.5, y: -2.5 },
          { x: 4.5, y: -2.5 },
          { x: 4.5, y: 2.5 },
          { x: -4.5, y: 2.5 },
          { x: -4.5, y: -2.5 }
        ]} />
      </footprint>
    }
    cadModel={{
      objUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/mbr1040.obj",
      stepUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/mbr1040.step",
      pcbRotationOffset: 90
    }}
    {...props}
  />
)