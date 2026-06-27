/**
 * PushButton - Tactile Switch Button
 * 
 * Standard 6x6mm tactile switch with through-hole pins
 * Momentary action - normally open, closes when pressed
 * 
 * Pins:
 * - 1: One side of switch
 * - 2: Other side of switch (internally connected to 1 when pressed)
 * - 3: Other side of switch  
 * - 4: Internally connected to 3 when pressed
 * 
 * For MCU connection, use pins 1 and 2 (or 3 and 4)
 * 
 * @category Controls
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const buttonPinLabels = {
  pin1: ["1", "COM1"],
  pin2: ["2", "NO1"],
  pin3: ["3", "COM2"],
  pin4: ["4", "NO2"]
} as const

export const PushButton = (props: ChipProps<typeof buttonPinLabels>) => (
  <chip
    pinLabels={buttonPinLabels}
    manufacturerPartNumber="SKRPACE010"
    supplierPartNumbers={{ jlcpcb: ["C121547"] }}
    footprint={
      <footprint>
        {/* Pin 1 - top left */}
        <platedhole portHints={["pin1"]} pcbX="-3.5mm" pcbY="-3.5mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 2 - top right */}
        <platedhole portHints={["pin2"]} pcbX="3.5mm" pcbY="-3.5mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 3 - bottom left */}
        <platedhole portHints={["pin3"]} pcbX="-3.5mm" pcbY="3.5mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 4 - bottom right */}
        <platedhole portHints={["pin4"]} pcbX="3.5mm" pcbY="3.5mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Button body outline */}
        <silkscreenpath route={[{ x: -5, y: -5 }, { x: 5, y: -5 }, { x: 5, y: 5 }, { x: -5, y: 5 }, { x: -5, y: -5 }]} />
        
        {/* Button center circle (actuator) */}
        <silkscreenpath route={[{ x: -2, y: -2 }, { x: 2, y: -2 }, { x: 2, y: 2 }, { x: -2, y: 2 }, { x: -2, y: -2 }]} />
        
        {/* Haptic bump indicator */}
        <silkscreentext text="●" pcbX="0mm" pcbY="0mm" fontSize="2mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="8mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -6, y: -6 },
          { x: 6, y: -6 },
          { x: 6, y: 6 },
          { x: -6, y: 6 },
          { x: -6, y: -6 }
        ]} />
      </footprint>
    }
    {...props}
  />
)
