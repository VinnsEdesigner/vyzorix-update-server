/**
 * LM5122 - Wide Vin Synchronous Boost Controller
 * 
 * High-efficiency synchronous boost controller for 12V to 40V conversion
 * Features: 3-65V Vin, external FETs, programmable UVLO, OCP, soft-start
 * 
 * Pins:
 * - VIN: Supply voltage
 * - GND: Ground
 * - SW: Switch node (MOSFET drive)
 * - HPFM: High-side FET drive
 * - HO: High-side gate drive output
 * - LO: Low-side gate drive output
 * - FB: Feedback (output voltage sense)
 * - COMP: Compensation pin
 * - RT: Timing resistor (sets frequency)
 * - EN: Enable input
 * - SS: Soft-start capacitor
 * - CS: Current sense input
 * 
 * @category Power
 * @package LM5122
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const lm5122PinLabels = {
  pin1: ["VIN"],
  pin2: ["VCC"],
  pin3: ["SW"],
  pin4: ["HPFM"],
  pin5: ["HO"],
  pin6: ["LO"],
  pin7: ["FB"],
  pin8: ["COMP"],
  pin9: ["RT"],
  pin10: ["EN"],
  pin11: ["SS"],
  pin12: ["CS"],
  pin13: ["AGND"],
  pin14: ["PGND"]
} as const

export const LM5122 = (props: ChipProps<typeof lm5122PinLabels>) => (
  <chip
    pinLabels={lm5122PinLabels}
    pinAttributes={{
      VIN: { requiresPower: true, mustBeConnected: true },
      VCC: { requiresPower: true, mustBeConnected: true },
      SW: { mustBeConnected: true },
      HPFM: { mustBeConnected: true },
      HO: { mustBeConnected: true },
      LO: { mustBeConnected: true },
      FB: { mustBeConnected: true },
      COMP: { mustBeConnected: true },
      RT: { mustBeConnected: true },
      EN: { mustBeConnected: true },
      SS: { mustBeConnected: true },
      CS: { mustBeConnected: true },
      AGND: { requiresGround: true },
      PGND: { requiresGround: true },
    }}
    manufacturerPartNumber="LM5122Q1MHX"
    supplierPartNumbers={{ jlcpcb: ["C94911"] }}
    footprint={
      <footprint>
        {/* Pin 1 - VIN */}
        <smtpad portHints={["pin1"]} pcbX="-3.2mm" pcbY="4.5mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 2 - VCC */}
        <smtpad portHints={["pin2"]} pcbX="-3.2mm" pcbY="3.2mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 3 - SW */}
        <smtpad portHints={["pin3"]} pcbX="-3.2mm" pcbY="1.9mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 4 - HPFM */}
        <smtpad portHints={["pin4"]} pcbX="-3.2mm" pcbY="0.6mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 5 - HO */}
        <smtpad portHints={["pin5"]} pcbX="-3.2mm" pcbY="-0.7mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 6 - LO */}
        <smtpad portHints={["pin6"]} pcbX="-3.2mm" pcbY="-2mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 7 - FB */}
        <smtpad portHints={["pin7"]} pcbX="-3.2mm" pcbY="-3.3mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 8 - COMP */}
        <smtpad portHints={["pin8"]} pcbX="3.2mm" pcbY="-3.3mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 9 - RT */}
        <smtpad portHints={["pin9"]} pcbX="3.2mm" pcbY="-2mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 10 - EN */}
        <smtpad portHints={["pin10"]} pcbX="3.2mm" pcbY="-0.7mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 11 - SS */}
        <smtpad portHints={["pin11"]} pcbX="3.2mm" pcbY="0.6mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 12 - CS */}
        <smtpad portHints={["pin12"]} pcbX="3.2mm" pcbY="1.9mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 13 - AGND */}
        <smtpad portHints={["pin13"]} pcbX="3.2mm" pcbY="3.2mm" width="0.7mm" height="1mm" shape="rect" />
        {/* Pin 14 - PGND */}
        <smtpad portHints={["pin14"]} pcbX="3.2mm" pcbY="4.5mm" width="0.7mm" height="1mm" shape="rect" />
        
        {/* Thermal pad (bottom center) */}
        <smtpad portHints={["pin13", "pin14"]} pcbX="0mm" pcbY="0mm" width="5mm" height="2.5mm" shape="rect" />
        
        {/* IC outline */}
        <silkscreenpath route={[{ x: -4, y: -5 }, { x: 4, y: -5 }, { x: 4, y: 5 }, { x: -4, y: 5 }, { x: -4, y: -5 }]} />
        
        {/* Pin 1 indicator (dot) */}
        <silkscreenpath route={[{ x: -4, y: 5 }, { x: -3.2, y: 5 }]} strokeWidth="0.5mm" />
        
        {/* LM5122 label */}
        <silkscreentext text="LM5122" pcbX="0mm" pcbY="6mm" fontSize="0.6mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-6.5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -5, y: -6 },
          { x: 5, y: -6 },
          { x: 5, y: 6 },
          { x: -5, y: 6 },
          { x: -5, y: -6 }
        ]} />
      </footprint>
    }
    {...props}
  />
)