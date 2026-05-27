export type DeviceStatus = "online" | "offline" | "warning" | "critical";

export interface MockDevice {
  id: string;
  name: string;
  model: string;
  androidVersion: string;
  status: DeviceStatus;
  riskScore: number;
  uptimeSec: number;
  thermalTemp: number;
  bufferLevel: number;
  appVersion: string;
  lastSeen: string;
}

export const mockDevices: MockDevice[] = [
  { id: "d-0001", name: "Nokia C22 — Lab A", model: "Nokia C22", androidVersion: "13", status: "online", riskScore: 12, uptimeSec: 184320, thermalTemp: 38.4, bufferLevel: 92, appVersion: "2.2.0", lastSeen: "2s ago" },
  { id: "d-0002", name: "Pixel 6a — QA", model: "Pixel 6a", androidVersion: "14", status: "online", riskScore: 28, uptimeSec: 90120, thermalTemp: 41.1, bufferLevel: 88, appVersion: "2.2.0", lastSeen: "5s ago" },
  { id: "d-0003", name: "Samsung A14 — Field 1", model: "SM-A145F", androidVersion: "13", status: "warning", riskScore: 61, uptimeSec: 23400, thermalTemp: 47.8, bufferLevel: 64, appVersion: "2.1.0", lastSeen: "11s ago" },
  { id: "d-0004", name: "Nokia C22 — Field 2", model: "Nokia C22", androidVersion: "13", status: "critical", riskScore: 82, uptimeSec: 6120, thermalTemp: 56.2, bufferLevel: 31, appVersion: "2.0.0", lastSeen: "1m ago" },
  { id: "d-0005", name: "Pixel 7 — Dev", model: "Pixel 7", androidVersion: "14", status: "offline", riskScore: 0, uptimeSec: 0, thermalTemp: 0, bufferLevel: 0, appVersion: "2.2.0", lastSeen: "2h ago" },
  { id: "d-0006", name: "Xiaomi Redmi 12", model: "Redmi 12", androidVersion: "13", status: "online", riskScore: 18, uptimeSec: 410220, thermalTemp: 39.7, bufferLevel: 95, appVersion: "2.2.0", lastSeen: "3s ago" },
  { id: "d-0007", name: "OnePlus Nord N20", model: "Nord N20", androidVersion: "12", status: "warning", riskScore: 54, uptimeSec: 51200, thermalTemp: 45.0, bufferLevel: 71, appVersion: "2.1.0", lastSeen: "18s ago" },
  { id: "d-0008", name: "Nokia G22 — Field 3", model: "Nokia G22", androidVersion: "13", status: "online", riskScore: 22, uptimeSec: 311000, thermalTemp: 40.2, bufferLevel: 89, appVersion: "2.2.0", lastSeen: "4s ago" },
];

export const mockAlerts = [
  { id: 1, severity: "critical" as const, device: "d-0004", message: "Risk score crossed 80 — soft reboot predicted", at: "12:04:31" },
  { id: 2, severity: "warning" as const, device: "d-0003", message: "Thermal sensor above 47°C — throttling engaged", at: "12:02:18" },
  { id: 3, severity: "warning" as const, device: "d-0007", message: "Buffer fill dropped below 75%", at: "11:58:02" },
  { id: 4, severity: "info" as const, device: "d-0001", message: "OTA check completed — up to date", at: "11:50:44" },
  { id: 5, severity: "info" as const, device: "d-0002", message: "Audio HAL reset acknowledged", at: "11:48:11" },
];

export const mockUpdateHistory = [
  { version: "2.2.0", versionCode: 220, releaseDate: "2026-05-20", fileSize: "12.4 MB", forced: false, downloads: 142 },
  { version: "2.1.0", versionCode: 210, releaseDate: "2026-04-08", fileSize: "12.1 MB", forced: false, downloads: 198 },
  { version: "2.0.0", versionCode: 200, releaseDate: "2026-02-14", fileSize: "11.7 MB", forced: true, downloads: 214 },
];

export const mockChangelog = [
  "Improved buffer health under thermal pressure on Nokia C22",
  "New RouteStateCard with last correction timestamp",
  "Reduced false-positive risk score on Pixel 7",
  "FCM wake reliability fixes for Doze mode",
];

export const mockLogLines = [
  "[12:04:31] [d-0004] AUDIOHAL: route loss detected, attempting REINIT_PROJECTION",
  "[12:04:29] [d-0004] RISK: score=82 trend=+12/min — predicting soft reboot",
  "[12:04:25] [d-0003] THERMAL: temp=47.8C policy=THROTTLE_LIGHT",
  "[12:04:18] [d-0001] TELEMETRY: uptime=184320 buffer=92 ok",
  "[12:04:12] [d-0002] CMD: TOGGLE_CAPTURE ack=true",
  "[12:04:05] [d-0007] BUFFER: fill=64 underrun=1",
  "[12:03:58] [d-0006] HEARTBEAT: ping/pong ok rtt=42ms",
  "[12:03:51] [d-0008] OTA: check=NOT_AVAILABLE current=2.2.0",
  "[12:03:44] [d-0001] CMD: FORCE_SPEAKER ack=true",
  "[12:03:30] [d-0005] CONN: dropped — entering FCM-wake mode",
];

export function makeSeries(n: number, base: number, variance: number) {
  return Array.from({ length: n }, (_, i) => ({
    t: i,
    value: Math.max(0, Math.round(base + (Math.random() - 0.5) * variance)),
  }));
}

export function formatUptime(sec: number) {
  if (!sec) return "—";
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}