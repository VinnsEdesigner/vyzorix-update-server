/**
 * PinHeader - 2.54mm Pitch Pin Header
 * 
 * Standard male pin header for I/O expansion
 * Can be configured for any number of pins (1x1 to 1xN)
 * 
 * @category Connectors
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const pinHeader4Labels = {
  pin1: ["1"],
  pin2: ["2"],
  pin3: ["3"],
  pin4: ["4"]
} as const

export const PinHeader4 = (props: ChipProps<typeof pinHeader4Labels>) => (
  <chip
    pinLabels={pinHeader4Labels}
    manufacturerPartNumber="2.54mm-1x4"
    supplierPartNumbers={{ jlcpcb: ["C49617"] }}
    footprint={
      <footprint>
        {/* Pin 1 */}
        <platedhole portHints={["pin1"]} pcbX="-4.75mm" pcbY="0mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 2 */}
        <platedhole portHints={["pin2"]} pcbX="-1.59mm" pcbY="0mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 3 */}
        <platedhole portHints={["pin3"]} pcbX="1.59mm" pcbY="0mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Pin 4 */}
        <platedhole portHints={["pin4"]} pcbX="4.75mm" pcbY="0mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        
        {/* Header outline */}
        <silkscreenpath route={[{ x: -6.5, y: -2 }, { x: 6.5, y: -2 }, { x: 6.5, y: 2 }, { x: -6.5, y: 2 }, { x: -6.5, y: -2 }]} />
        
        {/* Pin number indicators */}
        <silkscreentext text="1" pcbX="-4.75mm" pcbY="3mm" fontSize="0.5mm" anchorAlignment="center" />
        <silkscreentext text="2" pcbX="-1.59mm" pcbY="3mm" fontSize="0.5mm" anchorAlignment="center" />
        <silkscreentext text="3" pcbX="1.59mm" pcbY="3mm" fontSize="0.5mm" anchorAlignment="center" />
        <silkscreentext text="4" pcbX="4.75mm" pcbY="3mm" fontSize="0.5mm" anchorAlignment="center" />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -7, y: -2.5 },
          { x: 7, y: -2.5 },
          { x: 7, y: 2.5 },
          { x: -7, y: 2.5 },
          { x: -7, y: -2.5 }
        ]} />
      </footprint>
    }
    {...props}
  />
)

// Single pin header
const pinHeader1Labels = {
  pin1: ["1"]
} as const

export const PinHeader1 = (props: ChipProps<typeof pinHeader1Labels>) => (
  <chip
    pinLabels={pinHeader1Labels}
    manufacturerPartNumber="2.54mm-1x1"
    supplierPartNumbers={{ jlcpcb: ["C49611"] }}
    footprint={
      <footprint>
        <platedhole portHints={["pin1"]} pcbX="0mm" pcbY="0mm" outerDiameter="1.2mm" holeDiameter="0.8mm" shape="circle" />
        <silkscreenpath route={[{ x: -1.5, y: -1.5 }, { x: 1.5, y: -1.5 }, { x: 1.5, y: 1.5 }, { x: -1.5, y: 1.5 }, { x: -1.5, y: -1.5 }]} />
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="3mm" fontSize="0.8mm" anchorAlignment="center" />
        <courtyardoutline outline={[
          { x: -2, y: -2 },
          { x: 2, y: -2 },
          { x: 2, y: 2 },
          { x: -2, y: 2 },
          { x: -2, y: -2 }
        ]} />
      </footprint>
    }
    {...props}
  />
)
