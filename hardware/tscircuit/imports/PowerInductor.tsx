/**
 * PowerInductor - 22µH High-Current Power Inductor
 * 
 * Shielded construction for boost converter energy storage
 * Isat: 15A+, DCR: <10mΩ
 * 
 * Pins:
 * - pin1: Primary terminal 1
 * - pin2: Primary terminal 2
 * 
 * @category Power
 * @package Inductor
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const inductorPinLabels = {
  pin1: ["1"],
  pin2: ["2"]
} as const

export const PowerInductor = (props: ChipProps<typeof inductorPinLabels> & {
  inductance?: string
}) => (
  <chip
    pinLabels={inductorPinLabels}
    manufacturerPartNumber="SRN8040-220M"
    supplierPartNumbers={{ jlcpcb: ["C84824"] }}
    footprint={
      <footprint>
        {/* Pin 1 pad */}
        <smtpad portHints={["pin1"]} pcbX="-6mm" pcbY="0mm" width="2.5mm" height="3mm" shape="rect" />
        
        {/* Pin 2 pad */}
        <smtpad portHints={["pin2"]} pcbX="6mm" pcbY="0mm" width="2.5mm" height="3mm" shape="rect" />
        
        {/* Inductor body outline */}
        <silkscreenpath route={[{ x: -8, y: -6 }, { x: 8, y: -6 }, { x: 8, y: 6 }, { x: -8, y: 6 }, { x: -8, y: -6 }]} />
        
        {/* Coil symbol - spiral hint */}
        <silkscreenpath route={[{ x: -5, y: -3 }, { x: -5, y: 3 }, { x: 5, y: 3 }, { x: 5, y: -3 }]} />
        
        {/* Inductance marking */}
        <silkscreentext text="22µH" pcbX="0mm" pcbY="8mm" fontSize="0.6mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-8mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -9, y: -7 },
          { x: 9, y: -7 },
          { x: 9, y: 7 },
          { x: -9, y: 7 },
          { x: -9, y: -7 }
        ]} />
      </footprint>
    }
    cadModel={{
      objUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/inductor-srn8040.obj",
      stepUrl: "https://raw.githubusercontent.com/tscircuit/tscircuit/main/assets/3d-models/inductor-srn8040.step",
      pcbRotationOffset: 0
    }}
    {...props}
  />
)