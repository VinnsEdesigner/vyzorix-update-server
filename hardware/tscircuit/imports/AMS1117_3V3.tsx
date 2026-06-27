/**
 * AMS1117-3.3 - 3.3V Fixed Voltage Regulator
 * 
 * Low dropout (LDO) voltage regulator providing stable 3.3V from USB 5V
 * Output current: 1A (with adequate thermal management)
 * 
 * Pins:
 * - pin1: VIN (Input voltage)
 * - pin2: VOUT (Regulated output)
 * - pin3: GND (Ground)
 * 
 * @category Power
 * @package AMS1117
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const ams1117PinLabels = {
  pin1: ["VIN", "IN", "INPUT"],
  pin2: ["VOUT", "OUT", "OUTPUT"],
  pin3: ["GND", "GROUND"]
} as const

export const AMS1117_3V3 = (props: ChipProps<typeof ams1117PinLabels>) => (
  <chip
    pinLabels={ams1117PinLabels}
    manufacturerPartNumber="AMS1117-3.3"
    supplierPartNumbers={{ jlcpcb: ["C6186"] }}
    footprint={
      <footprint>
        {/* Input pin (left) */}
        <smtpad portHints={["pin1"]} pcbX="-2.5mm" pcbY="0mm" width="1mm" height="0.8mm" shape="rect" />
        
        {/* Output pin (center) */}
        <smtpad portHints={["pin2"]} pcbX="0mm" pcbY="0mm" width="1mm" height="0.8mm" shape="rect" />
        
        {/* Ground pin (right) */}
        <smtpad portHints={["pin3"]} pcbX="2.5mm" pcbY="0mm" width="1mm" height="0.8mm" shape="rect" />
        
        {/* Thermal pad (bottom) */}
        <smtpad portHints={["pin3"]} pcbX="0mm" pcbY="1.5mm" width="4mm" height="2mm" shape="rect" />
        
        {/* SOT-223 outline */}
        <silkscreenpath route={[{ x: -3.5, y: -1.5 }, { x: 3.5, y: -1.5 }, { x: 3.5, y: 1.5 }, { x: -3.5, y: 1.5 }, { x: -3.5, y: -1.5 }]} />
        
        {/* Pin 1 indicator (notch/tab) */}
        <silkscreenpath route={[{ x: -3.5, y: -1.5 }, { x: -2.5, y: -1.5 }]} />
        
        {/* 3.3V label */}
        <silkscreentext text="3.3V" pcbX="0mm" pcbY="3mm" fontSize="0.6mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-3mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -4, y: -2 },
          { x: 4, y: -2 },
          { x: 4, y: 3.5 },
          { x: -4, y: 3.5 },
          { x: -4, y: -2 }
        ]} />
      </footprint>
    }
    {...props}
  />
)
