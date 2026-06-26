import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"
import { PushButton } from "@tsci/seveibar.push-button"

/**
 * Vyzorix IoT Device Module
 * 
 * A production-grade IoT device featuring:
 * - USB-C power delivery
 * - Status LED indicator (blue)
 * - Reset and Boot buttons
 * - Decoupling capacitors for power integrity
 * - Pin header for external microcontroller connection
 * 
 * Board dimensions: 50mm x 40mm
 * 
 * @author @vinnsedesigner
 */
export default () => {
  return (
    <board width="50mm" height="40mm">
      {/* USB-C Power Connector */}
      <SmdUsbC
        name="USB_C"
        connections={{
          GND1: "net.GND",
          GND2: "net.GND",
          VBUS1: "net.VBUS",
          VBUS2: "net.VBUS",
        }}
        pcbX={-18}
        pcbY={8}
        schX={-15}
        schY={0}
        pinLabels={{
          GND1: { pinAttributes: ["require_ground"] },
          GND2: { pinAttributes: ["require_ground"] },
          VBUS1: { pinAttributes: ["require_power"] },
          VBUS2: { pinAttributes: ["require_power"] },
        }}
      />

      {/* Power Nets */}
      <net name="VBUS" />
      <net name="GND" />

      {/* Decoupling Capacitor - 100nF for high frequency filtering */}
      <capacitor
        name="C1"
        footprint="0603"
        capacitance="100nF"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={-8}
        pcbY={5}
        schX={-5}
        schY={5}
      />
      <trace from=".C1 .pos" to="net.VBUS" />
      <trace from=".C1 .neg" to="net.GND" />

      {/* Bulk decoupling capacitor - 10uF for low frequency */}
      <capacitor
        name="C2"
        footprint="0603"
        capacitance="10uF"
        supplierPartNumbers={{ jlcpcb: ["C28346"] }}
        pcbX={-8}
        pcbY={-2}
        schX={-5}
        schY={-5}
      />
      <trace from=".C2 .pos" to="net.VBUS" />
      <trace from=".C2 .neg" to="net.GND" />

      {/* Status LED with Current Limiting Resistor */}
      <resistor
        name="R1"
        footprint="0603"
        resistance="330R"
        supplierPartNumbers={{ jlcpcb: ["C25104"] }}
        pcbX={8}
        pcbY={8}
        schX={10}
        schY={5}
      />
      <led
        name="LED_STATUS"
        footprint="0603"
        ledColor="blue"
        supplierPartNumbers={{ jlcpcb: ["C433"] }}
        pcbX={15}
        pcbY={8}
        schX={15}
        schY={5}
      />
      {/* LED circuit: VBUS -> R1 -> LED -> GND */}
      <trace from=".R1 .pin1" to="net.VBUS" />
      <trace from=".R1 .pin2" to=".LED_STATUS .pos" />
      <trace from=".LED_STATUS .neg" to="net.GND" />

      {/* Reset Button - Active low, pulls EN to GND */}
      <PushButton
        name="BTN_RESET"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={-5}
        pcbY={-10}
        schX={-5}
        schY={-10}
        pinLabels={{ pin1: "EN", pin2: "GND" }}
      />
      <trace from=".BTN_RESET .pin1" to="net.GND" />
      <trace from=".BTN_RESET .pin2" to="net.GND" />

      {/* Boot Button - Used for boot mode selection */}
      <PushButton
        name="BTN_BOOT"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={5}
        pcbY={-10}
        schX={5}
        schY={-10}
        pinLabels={{ pin1: "GPIO0", pin2: "GND" }}
      />
      <trace from=".BTN_BOOT .pin1" to="net.GND" />
      <trace from=".BTN_BOOT .pin2" to="net.GND" />

      {/* Pin Header for external MCU/programming connections */}
      <chip
        name="JP1"
        footprint="pin_header_1x4"
        manufacturerPartNumber="2.54mm-1x4"
        pcbX={15}
        pcbY={-5}
        schX={15}
        schY={-8}
        pinLabels={{ pin1: "VUSB", pin2: "TX", pin3: "RX", pin4: "GND" }}
      />
    </board>
  )
}
