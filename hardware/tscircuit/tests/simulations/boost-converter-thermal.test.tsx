/**
 * Boost Converter Thermal Analysis Tests
 * 
 * Tests thermal performance of the 12V→40V @ 10A boost converter:
 * - MOSFET junction temperatures
 * - Thermal resistance calculations
 * - Heat sink requirements
 * - Safe operating area (SOA) verification
 * 
 * @category Simulation
 * @author VinnsEdesigner
 */

import { test, expect } from "vitest"

const AMBIENT_TEMP = 25 // °C
const MAX_JUNCTION_TEMP = 150 // °C (typical MOSFET limit)
const THERMAL_RESISTANCE_JA_CSD18537 = 50 // °C/W (junction to ambient for TO-263)
const THERMAL_RESISTANCE_JA_SOT23 = 200 // °C/W (for gate drivers)

test("MOSFET Q1 (High-Side) - Junction Temperature", async () => {
  /**
   * Thermal Analysis for High-Side MOSFET
   * 
   * At 400W output with synchronous boost:
   * - Q1 conducts during ~60% of cycle (off time)
   * - RMS current through Q1 ≈ 10A * sqrt(0.6) ≈ 7.75A
   * 
   * Power dissipation in Q1:
   * - Conduction: I_RMS² * R_DS(on) = 7.75² * 0.0023 ≈ 0.14W
   * - Switching: P_SW = 0.5 * V_IN * I_OUT * t_RISE * f_SW
   *            = 0.5 * 12 * 10 * 20ns * 200kHz ≈ 0.24W
   * - Total: ~0.38W
   */

  const i_rms = 7.75 // A
  const r_ds_on = 0.0023 // Ω (CSD18537NQ5A)
  const conduction_loss = i_rms * i_rms * r_ds_on
  
  const v_in = 12 // V
  const i_out = 10 // A
  const t_rise = 20e-9 // s (20ns)
  const f_sw = 200e3 // Hz (200kHz switching frequency)
  const switching_loss = 0.5 * v_in * i_out * t_rise * f_sw
  
  const total_loss_q1 = conduction_loss + switching_loss
  
  const junction_temp_q1 = AMBIENT_TEMP + (total_loss_q1 * THERMAL_RESISTANCE_JA_CSD18537)
  
  console.log(`Q1 Total Loss: ${total_loss_q1.toFixed(3)}W`)
  console.log(`Q1 Junction Temp: ${junction_temp_q1.toFixed(1)}°C`)
  
  expect(junction_temp_q1).toBeLessThan(MAX_JUNCTION_TEMP)
  expect(junction_temp_q1).toBeLessThan(100) // Keep well below max for reliability
})

test("MOSFET Q2 (Low-Side) - Junction Temperature", async () => {
  /**
   * Thermal Analysis for Low-Side MOSFET
   * 
   * At 400W output with synchronous boost:
   * - Q2 conducts during ~40% of cycle (on time)
   * - RMS current through Q2 ≈ 10A * sqrt(0.4) ≈ 6.32A
   * 
   * Power dissipation in Q2:
   * - Conduction: I_RMS² * R_DS(on) = 6.32² * 0.0023 ≈ 0.092W
   * - Switching: Similar to Q1
   * - Total: ~0.33W
   */

  const i_rms = 6.32 // A
  const r_ds_on = 0.0023 // Ω
  const conduction_loss = i_rms * i_rms * r_ds_on
  
  const v_in = 12 // V
  const i_out = 10 // A
  const t_rise = 20e-9 // s
  const f_sw = 200e3 // Hz
  const switching_loss = 0.5 * v_in * i_out * t_rise * f_sw
  
  const total_loss_q2 = conduction_loss + switching_loss
  
  const junction_temp_q2 = AMBIENT_TEMP + (total_loss_q2 * THERMAL_RESISTANCE_JA_CSD18537)
  
  console.log(`Q2 Total Loss: ${total_loss_q2.toFixed(3)}W`)
  console.log(`Q2 Junction Temp: ${junction_temp_q2.toFixed(1)}°C`)
  
  expect(junction_temp_q2).toBeLessThan(MAX_JUNCTION_TEMP)
  expect(junction_temp_q2).toBeLessThan(100)
})

test("Current Sense Resistor R_CS - Power Dissipation", async () => {
  /**
   * R_CS Thermal Analysis
   * 
   * Current sense resistor (20mΩ) carries the full inductor current:
   * - Average current: 10A (continuous)
   * - Power: I² * R = 10² * 0.02 = 2W
   * 
   * With 5W rating (2512 footprint), derating factor:
   * - 2W / 5W = 40% utilization - Safe
   */

  const i_sense = 10 // A (continuous)
  const r_cs = 0.02 // Ω
  const power_r_cs = i_sense * i_sense * r_cs
  
  const r_cs_rating = 5 // W (2512 footprint)
  const utilization = power_r_cs / r_cs_rating
  
  console.log(`R_CS Power: ${power_r_cs.toFixed(2)}W`)
  console.log(`R_CS Utilization: ${(utilization * 100).toFixed(1)}%`)
  
  expect(utilization).toBeLessThan(0.7) // Keep below 70% for reliability
  expect(power_r_cs).toBeLessThan(r_cs_rating)
})

test("Gate Driver Transistors Q3-Q6 - Thermal Check", async () => {
  /**
   * Gate Driver Thermal Analysis
   * 
   * Totem-pole drivers (MMBT2222A / MMBT3906):
   * - Average gate current: negligible (charge/discharge only)
   * - Peak current during switching: ~100mA
   * - Duty cycle: very low (~10%)
   * 
   * Power dissipation is minimal for gate drive only
   */

  const v_cc = 12 // V
  const v_gs = 10 // V (gate drive voltage)
  const q_g = 20e-9 // C (gate charge ~20nC for MOSFET)
  const f_sw = 200e3 // Hz
  
  // Average gate drive power (both drivers combined)
  const p_gate_drive = 2 * v_gs * q_g * f_sw // Both HS and LS
  
  // SOT23 thermal resistance
  const r_th_ja_sot23 = THERMAL_RESISTANCE_JA_SOT23
  
  const temp_rise = p_gate_drive * r_th_ja_sot23
  const junction_temp_drivers = AMBIENT_TEMP + temp_rise
  
  console.log(`Gate Drive Power: ${(p_gate_drive * 1000).toFixed(3)}mW`)
  console.log(`Driver Temp Rise: ${temp_rise.toFixed(3)}°C`)
  console.log(`Driver Junction Temp: ${junction_temp_drivers.toFixed(1)}°C`)
  
  // Gate drive power is negligible
  expect(p_gate_drive).toBeLessThan(0.1) // Less than 100mW
  expect(junction_temp_drivers).toBeLessThan(50) // Well within limits
})

test("Inductor L1 - Temperature Rise", async () =>
{
  /**
   * Inductor Thermal Analysis
   * 
   * Power inductor (22µH):
   * - DC resistance (DCR): ~10mΩ typical
   * - Current waveform: triangular with 10A peak
   * - RMS current: ~7A
   * 
   * Power loss: I_RMS² * DCR = 7² * 0.01 = 0.49W
   * Plus core losses: ~0.5W at 200kHz
   */

  const i_rms_inductor = 7 // A
  const dcr = 0.01 // Ω (typical for SRN8040)
  const copper_loss = i_rms_inductor * i_rms_inductor * dcr
  const core_loss = 0.5 // W (estimated for 22µH at 200kHz)
  
  const total_inductor_loss = copper_loss + core_loss
  
  // Inductor thermal resistance (varies by size)
  const r_th_inductor = 40 // °C/W for SRN8040 type
  
  const temp_rise_inductor = total_inductor_loss * r_th_inductor
  const inductor_hot_spot_temp = AMBIENT_TEMP + temp_rise_inductor
  
  console.log(`Inductor Copper Loss: ${copper_loss.toFixed(3)}W`)
  console.log(`Inductor Core Loss: ${core_loss.toFixed(3)}W`)
  console.log(`Inductor Temp Rise: ${temp_rise_inductor.toFixed(1)}°C`)
  console.log(`Inductor Hot Spot: ${inductor_hot_spot_temp.toFixed(1)}°C`)
  
  // Inductor rated for 125°C typically
  expect(inductor_hot_spot_temp).toBeLessThan(105) // 80°C rise from 25°C
})

test("Total System Thermal Budget", async () =>
{
  /**
   * System-Level Thermal Analysis
   * 
   * Total losses at 400W output:
   * - Q1 (HS MOSFET): ~0.38W
   * - Q2 (LS MOSFET): ~0.33W
   * - R_CS: ~2W
   * - L1: ~1W
   * - D1 (Schottky): ~0.5W (if used)
   * - Controller (LM5122): ~0.5W
   * - Gate drive losses: ~0.01W
   * 
   * Total: ~4.7W
   * 
   * PCB thermal design:
   * - 2oz copper (70µm) provides ~30°C/W per square
   * - With adequate copper area, system stays cool
   */

  const losses = {
    q1_mosfet: 0.38,
    q2_mosfet: 0.33,
    r_cs: 2.0,
    inductor: 1.0,
    schottky_d1: 0.5,
    controller: 0.5,
    gate_drive: 0.01,
  }
  
  const total_loss = Object.values(losses).reduce((a, b) => a + b, 0)
  
  // PCB copper area for heat spreading (100mm x 80mm board)
  // With 2oz copper, thermal resistance ~10°C/W
  const pcb_thermal_resistance = 10 // °C/W
  
  const system_temp_rise = total_loss * pcb_thermal_resistance
  const system_max_temp = AMBIENT_TEMP + system_temp_rise
  
  console.log("Thermal Loss Breakdown:")
  Object.entries(losses).forEach(([name, loss]) => {
    console.log(`  ${name}: ${loss.toFixed(2)}W`)
  })
  console.log(`Total System Loss: ${total_loss.toFixed(2)}W`)
  console.log(`System Temp Rise: ${system_temp_rise.toFixed(1)}°C`)
  console.log(`System Max Temperature: ${system_max_temp.toFixed(1)}°C`)
  
  expect(total_loss).toBeLessThan(10) // Keep total losses under 10W
  expect(system_max_temp).toBeLessThan(85) // Keep system below 85°C
})