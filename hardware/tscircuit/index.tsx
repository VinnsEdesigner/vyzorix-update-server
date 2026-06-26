import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"
import { PushButton } from "@tsci/seveibar.push-button"

/**
 * Vyzorix IoT Device Module
 * 
 * Production-grade IoT device module featuring:
 * - USB-C power delivery with proper decoupling
 * - Status LED with current limiting resistor
 * - Reset and Boot buttons for ESP32-style MCUs
 * - Pin header for programming/debug connections
 * 
 * Board: 50mm x 40mm
 * JLCPCB Production Ready
 * 
 * @author @vinnsedesigner
 */
export default () => {
  return (
    <board width="50mm" height="40mm">
      {/* Power Nets */}
      <net name="VBUS" />
      <net name="GND" />

      {/* ============================================ */}
      {/* USB-C Power Connector                       */}
      {/* ============================================ */}
      <SmdUsbC
        name="USB_C"
        pcbX={-15}
        pcbY={8}
        schX={-12}
        schY={0}
      />
      {/* USB-C Connections */}
      <trace from=".USB_C .GND1" to="net.GND" />
      <trace from=".USB_C .GND2" to="net.GND" />
      <trace from=".USB_C .VBUS1" to="net.VBUS" />
      <trace from=".USB_C .VBUS2" to="net.VBUS" />

      {/* ============================================ */}
      {/* High Frequency Decoupling - 100nF          */}
      {/* ============================================ */}
      <capacitor
        name="C1"
        footprint="0603"
        capacitance="100nF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={-8}
        pcbY={10}
        schX={-5}
        schY={8}
      />
      <trace from=".C1 .pos" to="net.VBUS" />
      <trace from=".C1 .neg" to="net.GND" />

      {/* ============================================ */}
      {/* Bulk Decoupling - 10uF                     */}
      {/* ============================================ */}
      <capacitor
        name="C2"
        footprint="0603"
        capacitance="10uF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C28346"] }}
        pcbX={-8}
        pcbY={5}
        schX={-5}
        schY={4}
      />
      <trace from=".C2 .pos" to="net.VBUS" />
      <trace from=".C2 .neg" to="net.GND" />

      {/* ============================================ */}
      {/* Status LED with Current Limiting            */}
      {/* ============================================ */}
      {/* 330R for ~10mA LED current */}
      <resistor
        name="R1"
        footprint="0603"
        resistance="330R"
        supplierPartNumbers={{ jlcpcb: ["C25104"] }}
        pcbX={5}
        pcbY={10}
        schX={5}
        schY={8}
      />
      <trace from=".R1 .pin1" to="net.VBUS" />

      <led
        name="LED_STATUS"
        footprint="0603"
        ledColor="blue"
        supplierPartNumbers={{ jlcpcb: ["C433"] }}
        pcbX={12}
        pcbY={10}
        schX={10}
        schY={8}
      />
      <trace from=".R1 .pin2" to=".LED_STATUS .pos" />
      <trace from=".LED_STATUS .neg" to="net.GND" />

      {/* ============================================ */}
      {/* Reset Button - Far bottom left              */}
      {/* ============================================ */}
      <PushButton
        name="BTN_RESET"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={-18}
        pcbY={-15}
        schX={-18}
        schY={-13}
      />
      <trace from=".BTN_RESET .pin1" to="net.GND" />
      <trace from=".BTN_RESET .pin2" to="net.GND" />

      {/* ============================================ */}
      {/* Boot Button - Far bottom right              */}
      {/* ============================================ */}
      <PushButton
        name="BTN_BOOT"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={18}
        pcbY={-15}
        schX={18}
        schY={-13}
      />
      <trace from=".BTN_BOOT .pin1" to="net.GND" />
      <trace from=".BTN_BOOT .pin2" to="net.GND" />

      {/* ============================================ */}
      {/* Programming Header - Right side             */}
      {/* ============================================ */}
      <chip
        name="JP1"
        footprint="dip_4"
        manufacturerPartNumber="1x4_2.54mm_header"
        pcbX={18}
        pcbY={-5}
        schX={18}
        schY={-3}
      />
      <trace from=".JP1 .pin4" to="net.GND" />
    </board>
  )
}
