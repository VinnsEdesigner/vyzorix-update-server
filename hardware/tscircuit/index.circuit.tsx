/**
 * 12V to 40V Boost Converter (Professional Design)
 * 
 * High-power synchronous boost converter with proper gate drivers
 * Input: 12V, Output: 40V @ 10A (400W)
 * 
 * Layout: Uses schematic sections for organized schematic view
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

export default () => (
  <board width="100mm" height="90mm" autorouterVersion="v4" layers={2}>
    
    {/* ======================================== */}
    {/* NETS */}
    {/* ======================================== */}
    <net name="GND" />
    <net name="VIN" />
    <net name="VOUT" />
    <net name="PHASE" />
    <net name="VCC" />
    <net name="BOOT" />
    <net name="ISENSE" />
    
    {/* ======================================== */}
    {/* POWER INPUT - LEFT */}
    {/* ======================================== */}
    <TerminalBlock2P name="VIN_P" pcbX={-44} pcbY={25} pcbRotation={90} />
    <TerminalBlock2P name="VIN_N" pcbX={-44} pcbY={-35} pcbRotation={90} />
    
    <capacitor name="C_IN1" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={-34} pcbY={25} />
    <capacitor name="C_IN2" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={-34} pcbY={17} />
    <capacitor name="C_IN3" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={-34} pcbY={9} />
    <capacitor name="C_BYP" capacitance="100nF" maxVoltageRating="50V" footprint="0603" {...jlcpcb("C14663")} pcbX={-34} pcbY={2} />
    
    <PowerInductor name="L1" inductance="22uH" pcbX={-15} pcbY={16} />
    
    {/* ======================================== */}
    {/* SWITCHING STAGE - CENTER */}
    {/* ======================================== */}
    <CSD18537NQ5A name="Q1" pcbX={0} pcbY={22} />
    <resistor name="R_G1" resistance="10R" footprint="0603" {...jlcpcb("C25804")} pcbX={8} pcbY={22} />
    <MBR1040 name="D1" pcbX={18} pcbY={22} pcbRotation={270} />
    
    <CSD18537NQ5A name="Q2" pcbX={0} pcbY={-22} />
    <resistor name="R_G2" resistance="10R" footprint="0603" {...jlcpcb("C25804")} pcbX={8} pcbY={-22} />
    <resistor name="R_CS" resistance="0.02R" tolerance="1%" footprint="2512" {...jlcpcb("C76748")} pcbX={0} pcbY={-38} />
    
    {/* ======================================== */}
    {/* GATE DRIVERS */}
    {/* ======================================== */}
    <transistor name="Q3" type="pnp" footprint="sot23" pcbX={14} pcbY={30} pcbRotation={90} />
    <transistor name="Q4" type="npn" footprint="sot23" pcbX={14} pcbY={20} pcbRotation={90} />
    <resistor name="R_PULLUP_HS" resistance="10k" footprint="0603" {...jlcpcb("C25804")} pcbX={20} pcbY={28} />
    
    <diode name="D_BOOT" pcbX={-10} pcbY={32} />
    <capacitor name="C_BOOT" capacitance="1uF" maxVoltageRating="50V" footprint="0805" {...jlcpcb("C14663")} pcbX={-18} pcbY={32} />
    
    <transistor name="Q5" type="pnp" footprint="sot23" pcbX={14} pcbY={-10} pcbRotation={90} />
    <transistor name="Q6" type="npn" footprint="sot23" pcbX={14} pcbY={-20} pcbRotation={90} />
    <resistor name="R_PULLUP_LS" resistance="10k" footprint="0603" {...jlcpcb("C25804")} pcbX={20} pcbY={-8} />
    
    {/* ======================================== */}
    {/* POWER OUTPUT - RIGHT */}
    {/* ======================================== */}
    <capacitor name="C_OUT1" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={35} pcbY={25} />
    <capacitor name="C_OUT2" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={35} pcbY={17} />
    <capacitor name="C_OUT3" capacitance="100uF" maxVoltageRating="50V" footprint="1206" {...jlcpcb("C19540")} pcbX={35} pcbY={9} />
    <capacitor name="C_FILT" capacitance="10uF" maxVoltageRating="50V" footprint="0805" {...jlcpcb("C14663")} pcbX={35} pcbY={2} />
    
    <TerminalBlock2P name="VOUT_P" pcbX={46} pcbY={25} pcbRotation={90} />
    <TerminalBlock2P name="VOUT_N" pcbX={46} pcbY={-35} pcbRotation={90} />
    
    {/* ======================================== */}
    {/* CONTROLLER - BOTTOM */}
    {/* ======================================== */}
    <LM5122 name="U1" pcbX={-25} pcbY={-30} />
    
    <capacitor name="C_SS" capacitance="10nF" maxVoltageRating="25V" footprint="0603" {...jlcpcb("C14663")} pcbX={-40} pcbY={-22} />
    <resistor name="R_RT" resistance="110k" footprint="0603" {...jlcpcb("C25804")} pcbX={-40} pcbY={-30} />
    <resistor name="R_FB1" resistance="33k" footprint="0603" {...jlcpcb("C25804")} pcbX={-12} pcbY={-22} />
    <resistor name="R_FB2" resistance="1k" footprint="0603" {...jlcpcb("C25804")} pcbX={-12} pcbY={-30} />
    <resistor name="R_COMP" resistance="10k" footprint="0603" {...jlcpcb("C25804")} pcbX={-20} pcbY={-44} />
    <capacitor name="C_COMP" capacitance="10nF" maxVoltageRating="25V" footprint="0603" {...jlcpcb("C14663")} pcbX={-12} pcbY={-44} />
    <resistor name="R_EN" resistance="10k" footprint="0603" {...jlcpcb("C25804")} pcbX={-38} pcbY={-44} />
    <capacitor name="C_EN" capacitance="1nF" maxVoltageRating="25V" footprint="0603" {...jlcpcb("C14663")} pcbX={-38} pcbY={-36} />
    <resistor name="R_LED_IN" resistance="1k" footprint="0603" {...jlcpcb("C25104")} pcbX={-30} pcbY={-44} />
    <led name="LED_IN" footprint="0805" {...jlcpcb("C83994")} pcbX={-30} pcbY={-36} />
    <resistor name="R_LED_OUT" resistance="3.3k" footprint="0603" {...jlcpcb("C25104")} pcbX={20} pcbY={-22} />
    <led name="LED_OUT" footprint="0805" {...jlcpcb("C83994")} pcbX={20} pcbY={-30} />
    
    {/* ======================================== */}
    {/* TRACES */}
    {/* ======================================== */}
    
    {/* VIN Power Path */}
    <trace from="VIN_P.pin1" to="C_IN1.pin1" />
    <trace from="C_IN1.pin1" to="C_IN2.pin1" />
    <trace from="C_IN2.pin1" to="C_IN3.pin1" />
    <trace from="C_IN3.pin1" to="C_BYP.pin1" />
    <trace from="C_BYP.pin1" to="net.VIN" />
    
    {/* GND Path */}
    <trace from="VIN_N.pin1" to="net.GND" />
    <trace from="VIN_N.pin2" to="net.GND" />
    <trace from="C_IN1.pin2" to="C_IN2.pin2" />
    <trace from="C_IN2.pin2" to="C_IN3.pin2" />
    <trace from="C_IN3.pin2" to="C_BYP.pin2" />
    <trace from="C_BYP.pin2" to="net.GND" />
    
    {/* VIN to VCC */}
    <trace from="net.VIN" to="net.VCC" />
    <trace from="net.VIN" to="L1.pin1" />
    
    {/* PHASE Node */}
    <trace from="L1.pin2" to="Q2.D" />
    <trace from="Q2.D" to="net.PHASE" />
    <trace from="net.PHASE" to="Q1.S" />
    <trace from="net.PHASE" to="D1.A" />
    
    {/* Current Sense */}
    <trace from="Q2.S" to="R_CS.pin1" />
    <trace from="R_CS.pin1" to="net.ISENSE" />
    <trace from="R_CS.pin2" to="net.GND" />
    
    {/* Output Node */}
    <trace from="Q1.D" to="D1.K" />
    <trace from="D1.K" to="C_OUT1.pin1" />
    <trace from="C_OUT1.pin1" to="C_OUT2.pin1" />
    <trace from="C_OUT2.pin1" to="C_OUT3.pin1" />
    <trace from="C_OUT3.pin1" to="C_FILT.pin1" />
    <trace from="C_FILT.pin1" to="net.VOUT" />
    
    {/* Output GND */}
    <trace from="C_OUT1.pin2" to="C_OUT2.pin2" />
    <trace from="C_OUT2.pin2" to="C_OUT3.pin2" />
    <trace from="C_OUT3.pin2" to="C_FILT.pin2" />
    <trace from="C_FILT.pin2" to="net.GND" />
    
    {/* Output Connectors */}
    <trace from="net.VOUT" to="VOUT_P.pin1" />
    <trace from="VOUT_P.pin2" to="VOUT_N.pin2" />
    <trace from="VOUT_N.pin1" to="net.GND" />
    
    {/* Bootstrap - diode anode to VOUT, cathode to boot cap */}
    <trace from="net.VOUT" to="D_BOOT.pin1" />
    <trace from="D_BOOT.pin2" to="C_BOOT.pin1" />
    <trace from="C_BOOT.pin1" to="net.BOOT" />
    <trace from="C_BOOT.pin2" to="net.GND" />
    
    {/* High-Side Gate Drive - Totem pole driver */}
    {/* Q3 is PNP: pin1=collector, pin2=emitter, pin3=base */}
    {/* Q4 is NPN: pin1=collector, pin3=base, pin2=emitter */}
    <trace from="U1.HO" to="Q3.pin3" />
    <trace from="Q3.pin2" to="net.BOOT" />
    <trace from="Q4.pin2" to="net.GND" />
    <trace from="Q3.pin1" to="R_PULLUP_HS.pin1" />
    <trace from="Q4.pin1" to="R_PULLUP_HS.pin1" />
    <trace from="R_PULLUP_HS.pin2" to="R_G1.pin1" />
    <trace from="R_G1.pin2" to="Q1.G" />
    
    {/* Low-Side Gate Drive - Totem pole driver */}
    {/* Q5 is PNP: pin1=collector, pin2=emitter, pin3=base */}
    {/* Q6 is NPN: pin1=collector, pin3=base, pin2=emitter */}
    <trace from="U1.LO" to="Q5.pin3" />
    <trace from="Q5.pin2" to="net.VCC" />
    <trace from="Q6.pin2" to="net.GND" />
    <trace from="Q5.pin1" to="R_PULLUP_LS.pin1" />
    <trace from="Q6.pin1" to="R_PULLUP_LS.pin1" />
    <trace from="R_PULLUP_LS.pin2" to="R_G2.pin1" />
    <trace from="R_G2.pin2" to="Q2.G" />
    
    {/* Switching Node Sense */}
    <trace from="U1.SW" to="net.PHASE" />
    <trace from="U1.HPFM" to="net.GND" />
    
    {/* Controller Power */}
    <trace from="net.VCC" to="U1.VIN" />
    <trace from="net.VCC" to="U1.VCC" />
    <trace from="U1.PGND" to="net.GND" />
    <trace from="U1.AGND" to="net.GND" />
    
    {/* Soft Start */}
    <trace from="U1.SS" to="C_SS.pin1" />
    <trace from="C_SS.pin2" to="net.GND" />
    
    {/* Timing */}
    <trace from="U1.RT" to="R_RT.pin1" />
    <trace from="R_RT.pin2" to="net.GND" />
    
    {/* Current Sense to Controller */}
    <trace from="U1.CS" to="net.ISENSE" />
    
    {/* Feedback */}
    <trace from="net.VOUT" to="R_FB1.pin1" />
    <trace from="R_FB1.pin2" to="R_FB2.pin1" />
    <trace from="R_FB2.pin2" to="net.GND" />
    <trace from="R_FB1.pin2" to="U1.FB" />
    
    {/* Compensation */}
    <trace from="U1.COMP" to="R_COMP.pin1" />
    <trace from="R_COMP.pin2" to="C_COMP.pin1" />
    <trace from="C_COMP.pin2" to="net.GND" />
    
    {/* Enable */}
    <trace from="U1.EN" to="R_EN.pin1" />
    <trace from="R_EN.pin2" to="net.VIN" />
    <trace from="R_EN.pin1" to="C_EN.pin1" />
    <trace from="C_EN.pin2" to="net.GND" />
    
    {/* Input LED */}
    <trace from="net.VIN" to="R_LED_IN.pin1" />
    <trace from="R_LED_IN.pin2" to="LED_IN.pin1" />
    <trace from="LED_IN.pin2" to="net.GND" />
    
    {/* Output LED */}
    <trace from="net.VOUT" to="R_LED_OUT.pin1" />
    <trace from="R_LED_OUT.pin2" to="LED_OUT.pin1" />
    <trace from="LED_OUT.pin2" to="net.GND" />
    
    {/* ======================================== */}
    {/* SILKSCREEN LABELS */}
    {/* ======================================== */}
    <pcbnotetext text="12V→40V BOOST CONVERTER" pcbX={0} pcbY={44} fontSize={2} anchorAlignment="center" />
    <pcbnotetext text="10A / 400W MAX" pcbX={0} pcbY={40} fontSize={1.2} anchorAlignment="center" />
    <pcbnotetext text="VinnsEdesigner 2026" pcbX={0} pcbY={-46} fontSize={0.7} anchorAlignment="center" />
  </board>
)
