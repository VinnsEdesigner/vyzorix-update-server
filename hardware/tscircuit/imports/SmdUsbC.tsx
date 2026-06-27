/**
 * SmdUsbC - Surface Mount USB Type-C Connector
 * 
 * Standard USB-C receptacle with through-hole mounting wings
 * 
 * Pins:
 * - pin1: GND
 * - pin2: VBUS
 * - pin3: CC1
 * - pin4: CC2
 * - pin5: DP
 * - pin6: DM
 * - pin7: SHIELD
 * 
 * @category Connectors
 * @author VinnsEdesigner
 */
import type { ChipProps } from "@tscircuit/props"

const usbCPinLabels = {
  pin1: ["GND", "G"],
  pin2: ["VBUS", "V"],
  pin3: ["CC1"],
  pin4: ["CC2"],
  pin5: ["D_P", "DP"],
  pin6: ["D_M", "DM"],
  pin7: ["SHIELD", "SH"]
} as const

export const SmdUsbC = (props: ChipProps<typeof usbCPinLabels>) => (
  <chip
    pinLabels={usbCPinLabels}
    manufacturerPartNumber="USB4085-GF"
    supplierPartNumbers={{ jlcpcb: ["C96551"] }}
    footprint={
      <footprint>
        {/* GND pad - top left */}
        <smtpad portHints={["pin1"]} pcbX="-1.8mm" pcbY="-2.5mm" width="0.6mm" height="1.2mm" shape="rect" />
        {/* VBUS pad - top */}
        <smtpad portHints={["pin2"]} pcbX="-1.8mm" pcbY="2.5mm" width="0.6mm" height="1.2mm" shape="rect" />
        
        {/* Shield wings - through hole style */}
        <platedhole portHints={["pin7"]} pcbX="-3.2mm" pcbY="0mm" outerDiameter="1.8mm" holeDiameter="1.2mm" shape="rect" />
        
        {/* USB Data lines */}
        <smtpad portHints={["pin5"]} pcbX="0.8mm" pcbY="-2.5mm" width="0.5mm" height="1mm" shape="rect" />
        <smtpad portHints={["pin6"]} pcbX="1.3mm" pcbY="-2.5mm" width="0.5mm" height="1mm" shape="rect" />
        
        {/* CC pins */}
        <smtpad portHints={["pin3"]} pcbX="0mm" pcbY="-2.5mm" width="0.5mm" height="1mm" shape="rect" />
        <smtpad portHints={["pin4"]} pcbX="0.5mm" pcbY="-2.5mm" width="0.5mm" height="1mm" shape="rect" />
        
        {/* USB-C outline silkscreen */}
        <silkscreenpath route={[{ x: -4.5, y: -3.5 }, { x: 4.5, y: -3.5 }, { x: 4.5, y: 3.5 }, { x: -4.5, y: 3.5 }, { x: -4.5, y: -3.5 }]} />
        
        {/* Center divider line */}
        <silkscreenpath route={[{ x: -2.5, y: -3.5 }, { x: -2.5, y: 3.5 }]} />
        
        {/* USB symbol hint */}
        <silkscreenpath route={[{ x: -0.5, y: -1.5 }, { x: 0.5, y: 0 }, { x: -0.5, y: 1.5 }]} />
        
        {/* Reference designator */}
        <silkscreentext text="{NAME}" pcbX="0mm" pcbY="-5mm" fontSize="0.8mm" anchorAlignment="center" />
        
        {/* Courtyard */}
        <courtyardoutline outline={[
          { x: -5, y: -4 },
          { x: 5, y: -4 },
          { x: 5, y: 4 },
          { x: -5, y: 4 },
          { x: -5, y: -4 }
        ]} />
      </footprint>
    }
    {...props}
  />
)
