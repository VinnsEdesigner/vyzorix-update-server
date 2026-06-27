import { SmdUsbC } from "@tsci/seveibar.smd-usb-c"

/**
 * Vyzorix Power Module - Production Schematic
 * 
 * Features:
 * - USB-C power input (5V/3A)
 * - 3.3V LDO regulation (AMS1117-3.3)
 * - Status LED indicators
 * - MCU interface header
 * 
 * Board: 30mm x 25mm
 * JLCPCB Basic Parts Assembly
 * 
 * @author @vinnsedesigner
 */
export default () => {
  return (
    <board width="30mm" height="25mm">
      {/* ============================================ */}
      {/* POWER RAIL BUSES                            */}
      {/* ============================================ */}
      <net name="VBUS" />
      <net name="V33" />
      <net name="GND" />

      {/* ============================================ */}
      {/* USB-C Power Input                           */}
      {/* ============================================ */}
      <SmdUsbC
        name="J1"
        pcbX={-10}
        pcbY={5}
        schX={-20}
        schY={-5}
      />
      <trace from=".J1 .VBUS1" to="net.VBUS" label="VBUS" />
      <trace from=".J1 .VBUS2" to="net.VBUS" label="VBUS" />
      <trace from=".J1 .GND1" to="net.GND" label="GND" />
      <trace from=".J1 .GND2" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* LDO Regulator U1 - 3.3V @ 800mA            */}
      {/* ============================================ */}
      <chip
        name="U1"
        footprint="SOT223"
        manufacturerPartNumber="AMS1117-3.3"
        pinLabels={{
          "1": ["VOUT"],
          "2": ["GND"],
          "3": ["VIN"],
          "4": ["GND"]
        }}
        pcbX={-2}
        pcbY={5}
        schX={-5}
        schY={-2}
      />
      <trace from=".U1 .VIN" to="net.VBUS" label="VIN" />
      <trace from=".U1 .VOUT" to="net.V33" label="VOUT" />
      <trace from=".U1 .GND" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* Input Capacitor C1 - 10uF                   */}
      {/* ============================================ */}
      <capacitor
        name="C1"
        footprint="0603"
        capacitance="10uF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C28346"] }}
        pcbX={-6}
        pcbY={2}
        schX={-10}
        schY={3}
      />
      <trace from=".C1 .pos" to="net.VBUS" label="VBUS" />
      <trace from=".C1 .neg" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* Output Capacitor C2 - 22uF                   */}
      {/* ============================================ */}
      <capacitor
        name="C2"
        footprint="0603"
        capacitance="22uF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C1154"] }}
        pcbX={2}
        pcbY={2}
        schX={2}
        schY={3}
      />
      <trace from=".C2 .pos" to="net.V33" label="V33" />
      <trace from=".C2 .neg" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* Decoupling Capacitor C3 - 100nF            */}
      {/* ============================================ */}
      <capacitor
        name="C3"
        footprint="0603"
        capacitance="100nF"
        voltageRating="16V"
        supplierPartNumbers={{ jlcpcb: ["C14663"] }}
        pcbX={5}
        pcbY={2}
        schX={8}
        schY={3}
      />
      <trace from=".C3 .pos" to="net.V33" label="V33" />
      <trace from=".C3 .neg" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* Status LED D1 - Power Indicator             */}
      {/* ============================================ */}
      <resistor
        name="R1"
        footprint="0603"
        resistance="330R"
        supplierPartNumbers={{ jlcpcb: ["C25104"] }}
        pcbX={10}
        pcbY={5}
        schX={12}
        schY={3}
      />
      <trace from=".R1 .pin1" to="net.V33" label="V33" />

      <led
        name="D1"
        footprint="0603"
        ledColor="green"
        supplierPartNumbers={{ jlcpcb: ["C433"] }}
        pcbX={13}
        pcbY={5}
        schX={16}
        schY={3}
      />
      <trace from=".D1 .pos" to=".R1 .pin2" />
      <trace from=".D1 .neg" to="net.GND" label="GND" />

      {/* ============================================ */}
      {/* MCU Interface Header J2 - 4-pin            */}
      {/* ============================================ */}
      <chip
        name="J2"
        footprint="dip_4"
        manufacturerPartNumber="CONN-4P-2.54mm"
        pinLabels={{
          "1": ["VBUS"],
          "2": ["V33"],
          "3": ["GND"],
          "4": ["DP"]
        }}
        pcbX={8}
        pcbY={-5}
        schX={15}
        schY={-3}
      />
      <trace from=".J2 .VBUS" to="net.VBUS" label="VBUS" />
      <trace from=".J2 .V33" to="net.V33" label="V33" />
      <trace from=".J2 .GND" to="net.GND" label="GND" />
      <trace from=".J2 .DP" to=".J1 .DP1" label="DP" />
      <trace from=".J1 .DM1" to=".J2 .pin4" label="DM" />
    </board>
  )
}
