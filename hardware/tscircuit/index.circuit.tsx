/**
 * 12V to 40V Boost Converter  Layout
 * 
 * High-power synchronous boost converter
 * Input: 12V, Output: 40V @ 10A (400W)
 * 
 * Layout: Wide spacing for clean routing
 * 
 * @category Power
 * @author VinnsEdesigner
 */

import { LM5122 } from "./imports/LM5122"
import { CSD18537NQ5A } from "./imports/CSD18537NQ5A"
import { MBR1040 } from "./imports/MBR1040"
import { PowerInductor } from "./imports/PowerInductor"
import { TerminalBlock2P } from "./imports/TerminalBlock2P"

const jlcpcb = (pn: string) => ({ supplierPartNumbers: { jlcpcb: [pn] } })
const p = (x: number, y: number, rotation = 0) => ({
  pcbX: x,
  pcbY: y,
  pcbRotation: rotation,
})

export default () => (
  <board width="100mm" height="80mm" autorouterVersion="v4" layers={2}>
    
    {/* ======================================== */}
    {/* NETS */}
    {/* ======================================== */}
    <net name="GND" />
    <net name="VIN" />
    <net name="VOUT" />
    <net name="PHASE" />
    <net name="VCC" />
    <net name="ISENSE" />
    
    {/* ======================================== */}
    {/* POWER INPUT - Far left */}
    {/* ======================================== */}
    
    <TerminalBlock2P name="VIN_P" {...p(-45, 30)} pcbRotation={90} />
    <TerminalBlock2P name="VIN_N" {...p(-45, -35)} pcbRotation={90} />
    
    <capacitor name="C_IN1" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(-35, 30)} />
    <capacitor name="C_IN2" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(-35, 25)} />
    <capacitor name="C_IN3" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(-35, 20)} />
    <capacitor name="C_BYP" capacitance="100nF" voltageRating="50V" footprint="0603" {...jlcpcb("C14663")} {...p(-35, 35)} />
    
    {/* ======================================== */}
    {/* POWER INDUCTOR */}
    {/* ======================================== */}
    <PowerInductor name="L1" inductance="22uH" {...p(-20, 25)} />
    
    {/* ======================================== */}
    {/* SWITCHING MOSFETS */}
    {/* ======================================== */}
    <CSD18537NQ5A name="Q1" {...p(-5, 25)} />
    <resistor name="R_G1" resistance="4.7R" footprint="0603" {...jlcpcb("C25804")} {...p(-5, 38)} />
    
    <CSD18537NQ5A name="Q2" {...p(12, 25)} />
    <resistor name="R_G2" resistance="4.7R" footprint="0603" {...jlcpcb("C25804")} {...p(12, 38)} />
    
    <resistor name="R_CS" resistance="0.02R" tolerance="1%" powerRating="5W" footprint="2512" {...jlcpcb("C76748")} {...p(3, 8)} />
    
    {/* ======================================== */}
    {/* OUTPUT DIODE & CAPS */}
    {/* ======================================== */}
    <MBR1040 name="D1" {...p(30, 25)} pcbRotation={90} />
    
    <capacitor name="C_OUT1" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(40, 30)} />
    <capacitor name="C_OUT2" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(40, 25)} />
    <capacitor name="C_OUT3" capacitance="100uF" voltageRating="50V" footprint="1206" {...jlcpcb("C19540")} {...p(40, 20)} />
    <capacitor name="C_FILT" capacitance="10uF" voltageRating="50V" footprint="0805" {...jlcpcb("C14663")} {...p(40, 35)} />
    
    <TerminalBlock2P name="VOUT_P" {...p(48, 30)} pcbRotation={90} />
    <TerminalBlock2P name="VOUT_N" {...p(48, -35)} pcbRotation={90} />
    
    {/* ======================================== */}
    {/* CONTROLLER IC */}
    {/* ======================================== */}
    <LM5122 name="U1" {...p(5, -20)} />
    
    <capacitor name="C_SS" capacitance="10nF" voltageRating="25V" footprint="0603" {...jlcpcb("C14663")} {...p(-8, -14)} />
    <resistor name="R_RT" resistance="110k" footprint="0603" {...jlcpcb("C25804")} {...p(-8, -28)} />
    
    <resistor name="R_FB1" resistance="33k" footprint="0603" {...jlcpcb("C25804")} {...p(20, -14)} />
    <resistor name="R_FB2" resistance="1k" footprint="0603" {...jlcpcb("C25804")} {...p(28, -14)} />
    <resistor name="R_COMP" resistance="10k" footprint="0603" {...jlcpcb("C25804")} {...p(20, -28)} />
    <capacitor name="C_COMP" capacitance="10nF" voltageRating="25V" footprint="0603" {...jlcpcb("C14663")} {...p(28, -28)} />
    
    <resistor name="R_EN" resistance="10k" footprint="0603" {...jlcpcb("C25804")} {...p(-15, -14)} />
    <capacitor name="C_EN" capacitance="1nF" voltageRating="25V" footprint="0603" {...jlcpcb("C14663")} {...p(-15, -28)} />
    
    <resistor name="R_LED_IN" resistance="1k" footprint="0603" {...jlcpcb("C25104")} {...p(-25, -14)} />
    <led name="LED_IN" footprint="0805" ledColor="green" {...jlcpcb("C83994")} {...p(-25, -6)} />
    
    <resistor name="R_LED_OUT" resistance="3.3k" footprint="0603" {...jlcpcb("C25104")} {...p(38, -14)} />
    <led name="LED_OUT" footprint="0805" ledColor="green" {...jlcpcb("C83994")} {...p(38, -6)} />
    
    {/* ======================================== */}
    {/* TRACES - POWER PATH */}
    {/* ======================================== */}
    <trace from="VIN_P.pin1" to="C_IN1.pin1" />
    <trace from="C_IN1.pin1" to="C_IN2.pin1" />
    <trace from="C_IN2.pin1" to="C_IN3.pin1" />
    <trace from="C_IN3.pin1" to="C_BYP.pin1" />
    <trace from="C_IN1.pin1" to="net.VIN" />
    
    <trace from="C_IN1.pin2" to="C_IN2.pin2" />
    <trace from="C_IN2.pin2" to="C_IN3.pin2" />
    <trace from="C_IN3.pin2" to="C_BYP.pin2" />
    <trace from="C_BYP.pin2" to="net.GND" />
    
    <trace from="VIN_N.pin1" to="net.GND" />
    <trace from="VIN_N.pin2" to="net.GND" />
    <trace from="net.VIN" to="net.VCC" />
    <trace from="net.VIN" to="L1.pin1" />
    
    <trace from="L1.pin2" to="Q1.D" />
    <trace from="Q1.D" to="net.PHASE" />
    <trace from="net.PHASE" to="Q2.D" />
    
    <trace from="Q1.S" to="R_CS.pin1" />
    <trace from="R_CS.pin1" to="net.ISENSE" />
    <trace from="R_CS.pin2" to="net.GND" />
    <trace from="Q2.S" to="net.GND" />
    
    <trace from="net.PHASE" to="D1.A" />
    
    <trace from="D1.K" to="C_OUT1.pin1" />
    <trace from="D1.K" to="C_OUT2.pin1" />
    <trace from="D1.K" to="C_OUT3.pin1" />
    <trace from="D1.K" to="C_FILT.pin1" />
    <trace from="D1.K" to="net.VOUT" />
    
    <trace from="C_OUT1.pin2" to="C_OUT2.pin2" />
    <trace from="C_OUT2.pin2" to="C_OUT3.pin2" />
    <trace from="C_OUT3.pin2" to="C_FILT.pin2" />
    <trace from="C_FILT.pin2" to="net.GND" />
    
    <trace from="net.VOUT" to="VOUT_P.pin1" />
    <trace from="VOUT_P.pin2" to="VOUT_N.pin2" />
    <trace from="VOUT_N.pin1" to="net.GND" />
    
    {/* ======================================== */}
    {/* TRACES - GATE DRIVE */}
    {/* ======================================== */}
    <trace from="U1.HO" to="R_G1.pin1" />
    <trace from="R_G1.pin2" to="Q1.G" />
    <trace from="U1.LO" to="R_G2.pin1" />
    <trace from="R_G2.pin2" to="Q2.G" />
    <trace from="U1.SW" to="net.PHASE" />
    <trace from="U1.HPFM" to="net.GND" />
    
    {/* ======================================== */}
    {/* TRACES - CONTROL */}
    {/* ======================================== */}
    <trace from="net.VCC" to="U1.VIN" />
    <trace from="net.VCC" to="U1.VCC" />
    <trace from="U1.PGND" to="net.GND" />
    <trace from="U1.AGND" to="net.GND" />
    <trace from="U1.SS" to="C_SS.pin1" />
    <trace from="C_SS.pin2" to="net.GND" />
    <trace from="U1.RT" to="R_RT.pin1" />
    <trace from="R_RT.pin2" to="net.GND" />
    <trace from="U1.CS" to="net.ISENSE" />
    
    <trace from="net.VOUT" to="R_FB1.pin1" />
    <trace from="R_FB1.pin2" to="R_FB2.pin1" />
    <trace from="R_FB2.pin2" to="net.GND" />
    <trace from="R_FB1.pin2" to="U1.FB" />
    <trace from="U1.COMP" to="R_COMP.pin1" />
    <trace from="R_COMP.pin2" to="C_COMP.pin1" />
    <trace from="C_COMP.pin2" to="net.GND" />
    
    <trace from="U1.EN" to="R_EN.pin1" />
    <trace from="R_EN.pin2" to="net.VIN" />
    <trace from="R_EN.pin1" to="C_EN.pin1" />
    <trace from="C_EN.pin2" to="net.GND" />
    
    <trace from="net.VIN" to="R_LED_IN.pin1" />
    <trace from="R_LED_IN.pin2" to="LED_IN.pin1" />
    <trace from="LED_IN.pin2" to="net.GND" />
    
    <trace from="net.VOUT" to="R_LED_OUT.pin1" />
    <trace from="R_LED_OUT.pin2" to="LED_OUT.pin1" />
    <trace from="LED_OUT.pin2" to="net.GND" />
    
    {/* ======================================== */}
    {/* SILKSCREEN */}
    {/* ======================================== */}
    <pcbnotetext text="12V→40V BOOST CONVERTER" pcbX={0} pcbY={42} fontSize={2} anchorAlignment="center" />
    <pcbnotetext text="10A / 400W" pcbX={0} pcbY={38} fontSize={1.2} anchorAlignment="center" />
    
    <pcbnotetext text="VIN 12V" pcbX={-45} pcbY={35} fontSize={0.8} anchorAlignment="center" />
    <pcbnotetext text="GND" pcbX={-45} pcbY={-40} fontSize={0.8} anchorAlignment="center" />
    <pcbnotetext text="VOUT 40V" pcbX={48} pcbY={35} fontSize={0.8} anchorAlignment="center" />
    <pcbnotetext text="GND" pcbX={48} pcbY={-40} fontSize={0.8} anchorAlignment="center" />
    
    <pcbnotetext text="POWER INPUT" pcbX={-35} pcbY={42} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="SWITCHING" pcbX={3} pcbY={42} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="OUTPUT" pcbX={40} pcbY={42} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="CONTROLLER" pcbX={5} pcbY={-10} fontSize={0.6} anchorAlignment="center" />
    
    <pcbnotetext text="L1" pcbX={-20} pcbY={32} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="22µH" pcbX={-20} pcbY={16} fontSize={0.5} anchorAlignment="center" />
    
    <pcbnotetext text="Q1" pcbX={-5} pcbY={32} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="HS" pcbX={-5} pcbY={16} fontSize={0.4} anchorAlignment="center" />
    
    <pcbnotetext text="Q2" pcbX={12} pcbY={32} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="LS" pcbX={12} pcbY={16} fontSize={0.4} anchorAlignment="center" />
    
    <pcbnotetext text="D1" pcbX={30} pcbY={32} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="SCHOTTKY" pcbX={30} pcbY={16} fontSize={0.35} anchorAlignment="center" />
    
    <pcbnotetext text="U1" pcbX={5} pcbY={-16} fontSize={0.6} anchorAlignment="center" />
    <pcbnotetext text="LM5122" pcbX={5} pcbY={-35} fontSize={0.5} anchorAlignment="center" />
    
    <pcbnotetext text="Rsense" pcbX={3} pcbY={4} fontSize={0.5} anchorAlignment="center" />
    <pcbnotetext text="0.02Ω" pcbX={3} pcbY={2} fontSize={0.4} anchorAlignment="center" />
    
    <pcbnotetext text="V1.0" pcbX={-45} pcbY={-40} fontSize={0.6} anchorAlignment="center" />
    
  </board>
)
