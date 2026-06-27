/**
 * TerminalBlock2P - 2-Position Screw Terminal Block
 * 
 * KF301-2P style terminal for power input/output connections
 * 5.08mm pitch, 15A rated
 * 
 * Pins:
 * - pin1: Terminal 1
 * - pin2: Terminal 2
 * 
 * @category Connectors
 * @package TerminalBlock
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const terminal2PinLabels = {
  pin1: ["1"],
  pin2: ["2"]
} as const

export const TerminalBlock2P = (props: ChipProps<typeof terminal2PinLabels>) => (
  <chip
    pinLabels={terminal2PinLabels}
    manufacturerPartNumber="KF301-2P"
    supplierPartNumbers={{ jlcpcb: ["C9176"] }}
    footprint={
      <footprint>
        {/* Terminal 1 - screw hole */}
        <platedhole portHints={["pin1"]} pcbX="-2.54mm" pcbY="0mm" outerDiameter="2mm" holeDiameter="1.2mm" shape="circle" />
        
        {/* Terminal 2 - screw hole */}
        <platedhole portHints={["pin2"]} pcbX="2.54mm" pcbY="0mm" outerDiameter="2mm" holeDiameter="1.2mm" shape="circle" />
        
        {/* Terminal block body outline */}
        <silkscreenpath route={[{ x: -5, y: -3 }, { x: 5, y: -3 }, { x: 5, y: 3 }, { x: -5, y: 3 }, { x: -5, y: -3 }]} />
        
        {/* Center divider */}
        <silkscreenpath route={[{ x: 0, y: -3 }, { x: 0, y: 3 }]} />
        
        {/* Screw slots */}
        <silkscreenpath route={[{ x: -2.54, y: -1.5 }, { x: -2.54, y: 1.5 }]} strokeWidth="0.3mm" />
        <silkscreenpath route={[{ x: 2.54, y: -1.5 }, { x: 2.54, y: 1.5 }]} strokeWidth="0.3mm" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -6, y: -4 },
          { x: 6, y: -4 },
          { x: 6, y: 4 },
          { x: -6, y: 4 },
          { x: -6, y: -4 }
        ]} />
      </footprint>
    }
    {...props}
  />
)