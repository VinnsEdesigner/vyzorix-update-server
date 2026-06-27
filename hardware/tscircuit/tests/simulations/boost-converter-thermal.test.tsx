/**
 * Boost Converter Thermal Analysis Tests
 * 
 * PRODUCTION REQUIREMENTS - Thermal limits for reliability
 * 
 * @category Simulation
 * @author VinnsEdesigner
 */

import { test, expect } from "vitest"

const AMBIENT_TEMP = 25 // °C
const MAX_JUNCTION_TEMP = 150 // °C (typical MOSFET limit)
const PRODUCTION_MAX_JUNCTION = 125 // °C (80% of max for reliability)
const THERMAL_RESISTANCE_JA_CSD18537 = 50 // °C/W (junction to ambient for TO-263)
const THERMAL_RESISTANCE_JA_SOT23 = 200 // °C/W (for gate drivers)

test("MOSFET Q1 (High-Side) - Junction Temperature", async () => {
  /**
   * PRODUCTION REQUIREMENT: Junction temp < 125°C (80% derating)
   * 
   * CSD18537NQ5A max junction: 150°C
   * Production limit: 125°C for 80% reliability margin
   */

  const i_rms = 7.75 // A (RMS current through HS MOSFET)
  const r_ds_on = 0.0023 // Ω (CSD18537NQ5A)
  const conduction_loss = i_rms * i_rms * r_ds_on
  
  const v_in = 12 // V
  const i_out = 10 // A
  const t_rise = 20e-9 // s (20ns)
  const f_sw = 90e3 // Hz (actual switching frequency)
  const switching_loss = 0.5 * v_in * i_out * t_rise * f_sw
  
  const total_loss_q1 = conduction_loss + switching_loss
  const junction_temp_q1 = AMBIENT_TEMP + (total_loss_q1 * THERMAL_RESISTANCE_JA_CSD18537)
  
  console.log(`Q1 Total Loss: ${total_loss_q1.toFixed(3)}W`)
  console.log(`Q1 Junction Temp: ${junction_temp_q1.toFixed(1)}°C`)
  console.log(`PRODUCTION MAX: ${PRODUCTION_MAX_JUNCTION}°C`)
  console.log(`STATUS: ${junction_temp_q1 <= PRODUCTION_MAX_JUNCTION ? 'PASS' : 'FAIL - Overheating'}`)
  
  // PRODUCTION THRESHOLD
  expect(junction_temp_q1).toBeLessThanOrEqual(PRODUCTION_MAX_JUNCTION)
})

test("MOSFET Q2 (Low-Side) - Junction Temperature", async () => {
  /**
   * PRODUCTION REQUIREMENT: Junction temp < 125°C
   */

  const i_rms = 6.32 // A
  const r_ds_on = 0.0023 // Ω
  const conduction_loss = i_rms * i_rms * r_ds_on
  
  const v_in = 12 // V
  const i_out = 10 // A
  const t_rise = 20e-9 // s
  const f_sw = 90e3 // Hz
  const switching_loss = 0.5 * v_in * i_out * t_rise * f_sw
  
  const total_loss_q2 = conduction_loss + switching_loss
  const junction_temp_q2 = AMBIENT_TEMP + (total_loss_q2 * THERMAL_RESISTANCE_JA_CSD18537)
  
  console.log(`Q2 Total Loss: ${total_loss_q2.toFixed(3)}W`)
  console.log(`Q2 Junction Temp: ${junction_temp_q2.toFixed(1)}°C`)
  console.log(`PRODUCTION MAX: ${PRODUCTION_MAX_JUNCTION}°C`)
  console.log(`STATUS: ${junction_temp_q2 <= PRODUCTION_MAX_JUNCTION ? 'PASS' : 'FAIL - Overheating'}`)
  
  // PRODUCTION THRESHOLD
  expect(junction_temp_q2).toBeLessThanOrEqual(PRODUCTION_MAX_JUNCTION)
})

test("Current Sense Resistor R_CS - Power Dissipation", async () => {
  /**
   * PRODUCTION REQUIREMENT: 50% derating for resistors
   * 
   * R_CS = 20mΩ, 5W rating (2512 footprint)
   * Production limit: 2.5W (50% derating)
   * Actual power at 10A: I² × R = 2W
   */

  const i_sense = 10 // A (continuous)
  const r_cs = 0.02 // Ω
  const power_r_cs = i_sense * i_sense * r_cs
  
  const r_cs_rating = 5 // W (2512 footprint)
  const production_limit = r_cs_rating * 0.5 // 50% derating
  
  console.log(`R_CS Power: ${power_r_cs.toFixed(2)}W`)
  console.log(`Resistor Rating: ${r_cs_rating}W`)
  console.log(`Production Limit (50% derating): ${production_limit}W`)
  console.log(`STATUS: ${power_r_cs <= production_limit ? 'PASS' : 'FAIL - Underrated'}`)
  
  // PRODUCTION THRESHOLD
  expect(power_r_cs).toBeLessThanOrEqual(production_limit)
})

test("Gate Driver Transistors Q3-Q6 - Thermal Check", async () => {
  /**
   * PRODUCTION REQUIREMENT: Junction temp < 100°C for SOT23 drivers
   */

  const v_gs = 10 // V (gate drive voltage)
  const q_g = 20e-9 // C (gate charge)
  const f_sw = 90e3 // Hz
  
  // Average gate drive power
  const p_gate_drive = 2 * v_gs * q_g * f_sw
  
  const temp_rise = p_gate_drive * THERMAL_RESISTANCE_JA_SOT23
  const junction_temp_drivers = AMBIENT_TEMP + temp_rise
  
  console.log(`Gate Drive Power: ${(p_gate_drive * 1000).toFixed(3)}mW`)
  console.log(`Driver Junction Temp: ${junction_temp_drivers.toFixed(1)}°C`)
  console.log(`PRODUCTION MAX: 100°C`)
  console.log(`STATUS: ${junction_temp_drivers <= 100 ? 'PASS' : 'FAIL'}`)
  
  // PRODUCTION THRESHOLD
  expect(junction_temp_drivers).toBeLessThanOrEqual(100)
})

test("Inductor L1 - Temperature Rise", async () => {
  /**
   * PRODUCTION REQUIREMENT: Inductor temp < 85°C (for 125°C rated parts)
   */

  const i_rms_inductor = 7 // A
  const dcr = 0.01 // Ω (typical for SRN8040)
  const copper_loss = i_rms_inductor * i_rms_inductor * dcr
  const core_loss = 0.5 // W (estimated)
  
  const total_inductor_loss = copper_loss + core_loss
  const r_th_inductor = 40 // °C/W
  const inductor_hot_spot_temp = AMBIENT_TEMP + (total_inductor_loss * r_th_inductor)
  
  console.log(`Inductor Total Loss: ${total_inductor_loss.toFixed(2)}W`)
  console.log(`Inductor Hot Spot: ${inductor_hot_spot_temp.toFixed(1)}°C`)
  console.log(`PRODUCTION MAX: 85°C`)
  console.log(`STATUS: ${inductor_hot_spot_temp <= 85 ? 'PASS' : 'FAIL'}`)
  
  // PRODUCTION THRESHOLD
  expect(inductor_hot_spot_temp).toBeLessThanOrEqual(85)
})

test("Total System Thermal Budget", async () => {
  /**
   * PRODUCTION REQUIREMENT: System must operate below 80°C ambient
   * 
   * DESIGN FIXED: Improved thermal design with larger board
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
  // Larger board 120x100mm with better copper
  const pcb_thermal_resistance = 10 // °C/W (improved thermal design)
  const system_max_temp = AMBIENT_TEMP + (total_loss * pcb_thermal_resistance)
  
  console.log(`Total System Loss: ${total_loss.toFixed(2)}W`)
  console.log(`PCB Thermal Resistance: ${pcb_thermal_resistance}°C/W`)
  console.log(`System Max Temperature: ${system_max_temp.toFixed(1)}°C`)
  console.log(`PRODUCTION MAX: 80°C ambient`)
  console.log(`STATUS: ${system_max_temp <= 80 ? 'PASS' : 'FAIL - Thermal issue'}`)
  
  // PRODUCTION THRESHOLD
  expect(system_max_temp).toBeLessThanOrEqual(80)
})