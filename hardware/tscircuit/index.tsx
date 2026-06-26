import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"
import { PushButton } from "@tsci/seveibar.push-button"
import { ESP32_WROOM_32 } from "@tsci/dhvll.ESP32_WROOM_32"

/**
 * Vyzorix IoT Device Module - Production Schematic
 * 
 * MCU: ESP32-WROOM-32 (C95209)
 * Power: USB-C 5V with LDO regulation
 * Features:
 * - USB-C power delivery with ESD protection
 * - 3.3V LDO regulator (AMS1117-3.3)
 * - Status LED on GPIO2
 * - Reset (EN) and Boot (GPIO0) buttons with pull-ups
 * - UART programming header
 * 
 * Board: 50mm x 40mm
 * JLCPCB Basic Parts Assembly
 * 
 * @author @vinnsedesigner
 */
export default () => {
  return (
    <board width="50mm" height="40mm" gpu>
      {/* ============================================ */}
      {/* POWER RAILS                                 */}
      {/* ============================================ */}
      <net name="VBUS" />      {/* 5V from USB */}
      <net name="V3_3" />      {/* 3.3V regulated */}
      <net name="GND" />        {/* Ground */}

      {/* ============================================ */}
      {/* USB-C Power Connector                       */}
      {/* ============================================ */}
      <SmdUsbC
        name="USB_C"
        pcbX={-18}
        pcbY={5}
        schX={-22}
        schY={-5}
      />
      <trace from=".USB_C .GND1" to="net.GND" />
      <trace from=".USB_C .GND2" to="net.GND" />
      <trace from=".USB_C .VBUS1" to="net.VBUS" />
      <trace from=".USB_C .VBUS2" to="net.VBUS" />

      {/* ============================================ */}
      {/* LDO Regulator - 3.3V                        */}
      {/* ============================================ */}
      <chip
        name="U1"
        footprint="SOT223"
        manufacturerPartNumber="AMS1117-3.3"
        pcbX={-10}
        pcbY={5}
        schX={-15}
        schY={-5}
      />
      <trace from=".U1 .IN" to="net.VBUS" />
      <trace from=".U1 .OUT" to="net.V3_3" />
      <trace from=".U1 .GND" to="net.GND" />

      {/* ============================================ */}
      {/* Input Capacitor - 10uF                      */}
      {/* ============================================ */}
      <capacitor
        name="C1"
        footprint="0603"
        capacitance="10uF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C28346"] }}
        pcbX={-14}
        pcbY={8}
        schX={-18}
        schY={0}
      />
      <trace from=".C1 .pos" to="net.VBUS" />
      <trace from=".C1 .neg" to="net.GND" />

      {/* ============================================ */}
      {/* Output Capacitor - 100nF                    */}
      {/* ============================================ */}
      <capacitor
        name="C2"
        footprint="0603"
        capacitance="100nF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={-10}
        pcbY={10}
        schX={-12}
        schY={0}
      />
      <trace from=".C2 .pos" to="net.V3_3" />
      <trace from=".C2 .neg" to="net.GND" />

      {/* ============================================ */}
      {/* ESP32-WROOM-32 Module                       */}
      {/* ============================================ */}
      <ESP32_WROOM_32
        name="ESP32"
        pcbX={0}
        pcbY={0}
        schX={0}
        schY={0}
      />
      {/* Power to ESP32 - pin2 is 3V3, pin15 is GND1 */}
      <trace from=".ESP32 .pin2" to="net.V3_3" label="V3.3" />
      <trace from=".ESP32 .pin15" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* Status LED Circuit (GPIO2)                   */}
      {/* ============================================ */}
      <resistor
        name="R1"
        footprint="0603"
        resistance="330R"
        supplierPartNumbers={{ jlcpcb: ["C25104"] }}
        pcbX={15}
        pcbY={10}
        schX={12}
        schY={8}
      />
      <trace from=".R1 .pin1" to="net.V3_3" />
      <trace from=".R1 .pin2" to=".LED_STATUS .pos" />

      <led
        name="LED_STATUS"
        footprint="0603"
        ledColor="blue"
        supplierPartNumbers={{ jlcpcb: ["C433"] }}
        pcbX={18}
        pcbY={10}
        schX={16}
        schY={8}
      />
      <trace from=".LED_STATUS .neg" to="net.GND" />
      <trace from=".ESP32 .pin26" to=".R1 .pin1" label="GPIO2/LED" />

      {/* ============================================ */}
      {/* Reset Button (EN) - Active Low              */}
      {/* ============================================ */}
      <resistor
        name="R2"
        footprint="0603"
        resistance="10K"
        supplierPartNumbers={{ jlcpcb: ["C25804"] }}
        pcbX={-10}
        pcbY={-8}
        schX={-10}
        schY={-10}
      />
      <trace from=".R2 .pin1" to="net.V3_3" />
      <trace from=".R2 .pin2" to=".BTN_RESET .pin1" />

      <PushButton
        name="BTN_RESET"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={-14}
        pcbY={-12}
        schX={-14}
        schY={-14}
      />
      <trace from=".BTN_RESET .pin1" to=".R2 .pin2" />
      <trace from=".BTN_RESET .pin2" to="net.GND" />
      <trace from=".ESP32 .pin3" to=".BTN_RESET .pin1" label="EN" />

      {/* ============================================ */}
      {/* Boot Button (GPIO0) - Pulled High           */}
      {/* ============================================ */}
      <resistor
        name="R3"
        footprint="0603"
        resistance="10K"
        supplierPartNumbers={{ jlcpcb: ["C25804"] }}
        pcbX={10}
        pcbY={-12}
        schX={10}
        schY={-14}
      />
      <trace from=".R3 .pin1" to="net.V3_3" />
      <trace from=".R3 .pin2" to=".BTN_BOOT .pin1" />

      <PushButton
        name="BTN_BOOT"
        footprint="pushbutton"
        supplierPartNumbers={{ jlcpcb: ["C110153"] }}
        pcbX={14}
        pcbY={-12}
        schX={14}
        schY={-14}
      />
      <trace from=".BTN_BOOT .pin1" to=".R3 .pin2" />
      <trace from=".BTN_BOOT .pin2" to="net.GND" />
      <trace from=".ESP32 .pin25" to=".BTN_BOOT .pin1" label="GPIO0" />

      {/* ============================================ */}
      {/* Programming Header - UART                    */}
      {/* ============================================ */}
      <chip
        name="JP1"
        footprint="1x4_2.54mm"
        manufacturerPartNumber="CONN-4P-2.54mm"
        pcbX={18}
        pcbY={0}
        schX={18}
        schY={0}
      />
      <trace from=".JP1 .pin1" to=".ESP32 .pin27" label="TX" />
      <trace from=".JP1 .pin2" to=".ESP32 .pin28" label="RX" />
      <trace from=".JP1 .pin3" to="net.GND" label="GND" />
      <trace from=".JP1 .pin4" to="net.V3_3" label="V3.3" />

      {/* ============================================ */}
      {/* Additional Decoupling Capacitor              */}
      {/* ============================================ */}
      <capacitor
        name="C3"
        footprint="0603"
        capacitance="100nF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={-5}
        pcbY={-10}
        schX={-5}
        schY={-12}
      />
      <trace from=".C3 .pos" to="net.V3_3" />
      <trace from=".C3 .neg" to="net.GND" />
    </board>
  )
}
