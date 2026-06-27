/**
 * CSD18537NQ5A - 40V 100A N-Channel MOSFET
 * 
 * Low Rds(on) power MOSFET for high-current boost converter applications
 * Vds: 40V, Id: 100A pulse, Rds(on): 3.2mΩ
 * Package: SON 5×6mm
 * 
 * Pins:
 * - G: Gate
 * - D: Drain
 * - S: Source
 * 
 * @category Power
 * @package CSD18537NQ5A
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const mosfetPinLabels = {
  pin1: ["G", "GATE"],
  pin2: ["D", "DRAIN"],
  pin3: ["S", "SOURCE"]
} as const

export const CSD18537NQ5A = (props: ChipProps<typeof mosfetPinLabels>) => (
  <chip
    pinLabels={mosfetPinLabels}
    pinAttributes={{
      G: { mustBeConnected: true },
      D: { requiresPower: true },
      S: { requiresPower: true },
    }}
    manufacturerPartNumber="CSD18537NQ5A"
    supplierPartNumbers={{ jlcpcb: ["C141250"] }}
    footprint={
      <footprint>
        {/* Gate pad (top center) */}
        <smtpad portHints={["pin1"]} pcbX="0mm" pcbY="4mm" width="0.8mm" height="0.8mm" shape="rect" />
        
        {/* Drain pad (left side - connects to inductor/PHASE) */}
        <smtpad portHints={["pin2"]} pcbX="-3.5mm" pcbY="-1mm" width="2.5mm" height="2mm" shape="rect" />
        
        {/* Source pad (right side - connects to GND or PHASE) */}
        <smtpad portHints={["pin3"]} pcbX="3.5mm" pcbY="-1mm" width="2.5mm" height="2mm" shape="rect" />
        
        {/* Thermal pad (bottom) */}
        <smtpad portHints={["pin2", "pin3"]} pcbX="0mm" pcbY="-4mm" width="6mm" height="2mm" shape="rect" />
        
        {/* MOSFET outline */}
        <silkscreenpath route={[{ x: -5, y: -5 }, { x: 5, y: -5 }, { x: 5, y: 5 }, { x: -5, y: 5 }, { x: -5, y: -5 }]} />
        
        {/* Pin 1 indicator */}
        <silkscreenpath route={[{ x: -5, y: 5 }, { x: -3, y: 5 }]} strokeWidth="0.4mm" />
        
        {/* MOSFET symbol - N-channel indicator */}
        <silkscreentext text="N" pcbX="0mm" pcbY="0mm" fontSize="1.5mm" anchorAlignment="center" />
        
        {/* Part label */}
        <silkscreentext text="CSD18537" pcbX="0mm" pcbY="6mm" fontSize="0.5mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-6.5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -6, y: -6.5 },
          { x: 6, y: -6.5 },
          { x: 6, y: 6.5 },
          { x: -6, y: 6.5 },
          { x: -6, y: -6.5 }
        ]} />
      </footprint>
    }
    cadModel={{
      objUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/csd18537NQ5A.obj",
      stepUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/csd18537NQ5A.step",
      pcbRotationOffset: 90
    }}
    {...props}
  />
)