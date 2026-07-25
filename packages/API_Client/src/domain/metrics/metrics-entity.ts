export type TimeRange = "1h" | "6h" | "24h" | "7d";

export type MetricResolution = "1m" | "5m" | "15m" | "1h";

export interface MetricChartPoint {
  timestamp: Date;
  value: number;
}

export interface MetricThreshold {
  warning: number;
  critical: number;
}

export interface MetricData {
  current: number;
  avg: number;
  min: number;
  max: number;
  unit: string;
  chart: MetricChartPoint[];
  threshold: MetricThreshold;
}

export interface UptimeData {
  current: number;
  unit: string;
}

export interface MetricEvents {
  timestamp: Date;
  type: string;
  metric: string;
  value: number;
  threshold: number;
}

export interface TimeRangeInfo {
  start: Date;
  end: Date;
  range: TimeRange;
  resolution: MetricResolution;
}

export interface DeviceBasic {
  imei: string;
  deviceName?: string;
}

export interface DeviceMetrics {
  device: DeviceBasic;
  timeRange: TimeRangeInfo;
  metrics: {
    riskScore: MetricData;
    thermalTemp: MetricData;
    bufferLevel: MetricData;
    uptime: UptimeData;
  };
  events: MetricEvents[];
}

export interface DashboardStats {
  devices: {
    total: number;
    online: number;
    offline: number;
  };
  commands: {
    totalToday: number;
    pending: number;
    failed: number;
  };
  activity: {
    last24h: {
      commands: number;
      registrations: number;
      deregistrations: number;
    };
  };
}

export function getTimeRangeMs(range: TimeRange): number {
  const ranges: Record<TimeRange, number> = {
    "1h": 60 * 60 * 1000,
    "6h": 6 * 60 * 60 * 1000,
    "24h": 24 * 60 * 60 * 1000,
    "7d": 7 * 24 * 60 * 60 * 1000,
  };
  return ranges[range];
}

export function getResolutionForRange(range: TimeRange): MetricResolution {
  const resolutions: Record<TimeRange, MetricResolution> = {
    "1h": "1m",
    "6h": "5m",
    "24h": "15m",
    "7d": "1h",
  };
  return resolutions[range];
}

export function getTimeRangeLabel(range: TimeRange): string {
  const labels: Record<TimeRange, string> = {
    "1h": "1 Hour",
    "6h": "6 Hours",
    "24h": "24 Hours",
    "7d": "7 Days",
  };
  return labels[range];
}
