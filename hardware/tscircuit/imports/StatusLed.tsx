/**
 * StatusLed - Status Indicator LED with Polarity Mark
 * 
 * Standard 0805/0603 LED with cathode bar marking
 * 
 * Pins:
 * - pin1: Anode (positive, connect to resistor)
 * - pin2: Cathode (negative, connect to GND via resistor)
 * 
 * @category Indicators
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const ledPinLabels = {
  pin1: ["A", "ANODE", "POS"],
  pin2: ["K", "CATHODE", "NEG"]
} as const

export const StatusLed = (props: ChipProps<typeof ledPinLabels> & {
  ledColor?: "red" | "green" | "blue" | "yellow" | "white"
}) => (
  <chip
    pinLabels={ledPinLabels}
    pinAttributes={{
      A: { requiresPower: true },
      K: { requiresPower: true },
    }}
    manufacturerPartNumber="LTST-C171"
    supplierPartNumbers={{ jlcpcb: ["C83994"] }}
    schWidth={1.35}
    footprint={
      <footprint>
        {/* Anode pad (left) */}
        <smtpad portHints={["pin1"]} pcbX="-1.5mm" pcbY="0mm" width="0.8mm" height="0.6mm" shape="rect" />
        
        {/* Cathode pad (right) */}
        <smtpad portHints={["pin2"]} pcbX="1.5mm" pcbY="0mm" width="0.8mm" height="0.6mm" shape="rect" />
        
        {/* LED body outline */}
        <silkscreenpath route={[{ x: -2.5, y: -1 }, { x: 2.5, y: -1 }, { x: 2.5, y: 1 }, { x: -2.5, y: 1 }, { x: -2.5, y: -1 }]} />
        
        {/* Diode symbol - anode side */}
        <silkscreenpath route={[{ x: -0.5, y: -0.7 }, { x: -0.5, y: 0.7 }]} />
        <silkscreenpath route={[{ x: -0.8, y: -0.7 }, { x: -0.5, y: 0 }, { x: -0.8, y: 0.7 }]} />
        
        {/* Diode symbol - cathode side */}
        <silkscreenpath route={[{ x: 0.5, y: -0.7 }, { x: 0.5, y: 0.7 }]} />
        <silkscreenpath route={[{ x: 0.5, y: -0.7 }, { x: 0.8, y: 0 }]} />
        <silkscreenpath route={[{ x: 0.5, y: 0.7 }, { x: 0.8, y: 0 }]} />
        
        {/* Cathode bar (flat side of LED) */}
        <silkscreenpath route={[{ x: 2.5, y: -1 }, { x: 2.5, y: 1 }]} strokeWidth="0.3mm" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-2.5mm" fontSize="0.6mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -3, y: -1.5 },
          { x: 3, y: -1.5 },
          { x: 3, y: 1.5 },
          { x: -3, y: 1.5 },
          { x: -3, y: -1.5 }
        ]} />
      </footprint>
    }
    {...props}
  />
)
