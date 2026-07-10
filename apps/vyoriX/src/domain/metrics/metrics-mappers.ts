import type {
  DeviceMetrics,
  DashboardStats,
  MetricChartPoint,
  MetricData,
  UptimeData,
  TimeRangeInfo,
  MetricEvents,
  TimeRange,
  MetricResolution,
} from "./metrics-entity";

export type RawMetricChartPoint = {
  timestamp: number;
  value: number;
};

export type RawMetricThreshold = {
  warning: number;
  critical: number;
};

export type RawMetricData = {
  current: number;
  avg: number;
  min: number;
  max: number;
  unit: string;
  chart: RawMetricChartPoint[];
  threshold?: RawMetricThreshold;
};

export type RawUptimeData = {
  current: number;
  unit: string;
};

export type RawMetricEvents = {
  timestamp: number;
  type: string;
  metric: string;
  value: number;
  threshold: number;
};

export type RawTimeRangeInfo = {
  start: number;
  end: number;
  range: string;
  resolution: string;
};

export type RawDeviceBasic = {
  imei: string;
  device_name?: string;
};

export type RawDeviceMetrics = {
  device: RawDeviceBasic;
  time_range: RawTimeRangeInfo;
  metrics: {
    risk_score: RawMetricData;
    thermal_temp: RawMetricData;
    buffer_level: RawMetricData;
    uptime: RawUptimeData;
  };
  events: RawMetricEvents[];
};

export type RawDashboardStats = {
  devices: {
    total: number;
    online: number;
    offline: number;
  };
  commands: {
    total_today: number;
    pending: number;
    failed: number;
  };
  activity: {
    last_24h: {
      commands: number;
      registrations: number;
      deregistrations: number;
    };
  };
};

function chartPointFromRaw(raw: RawMetricChartPoint): MetricChartPoint {
  return {
    timestamp: new Date(raw.timestamp),
    value: raw.value,
  };
}

function metricDataFromRaw(raw: RawMetricData): MetricData {
  return {
    current: raw.current,
    avg: raw.avg,
    min: raw.min,
    max: raw.max,
    unit: raw.unit,
    chart: raw.chart.map(chartPointFromRaw),
    threshold: raw.threshold,
  };
}

function uptimeFromRaw(raw: RawUptimeData): UptimeData {
  return {
    current: raw.current,
    unit: raw.unit,
  };
}

function metricEventFromRaw(raw: RawMetricEvents): MetricEvents {
  return {
    timestamp: new Date(raw.timestamp),
    type: raw.type,
    metric: raw.metric,
    value: raw.value,
    threshold: raw.threshold,
  };
}

function timeRangeFromRaw(raw: RawTimeRangeInfo): TimeRangeInfo {
  return {
    start: new Date(raw.start),
    end: new Date(raw.end),
    range: raw.range as TimeRange,
    resolution: raw.resolution as MetricResolution,
  };
}

export function deviceMetricsFromRaw(raw: RawDeviceMetrics): DeviceMetrics {
  return {
    device: {
      imei: raw.device.imei,
      deviceName: raw.device.device_name,
    },
    timeRange: timeRangeFromRaw(raw.time_range),
    metrics: {
      riskScore: metricDataFromRaw(raw.metrics.risk_score),
      thermalTemp: metricDataFromRaw(raw.metrics.thermal_temp),
      bufferLevel: metricDataFromRaw(raw.metrics.buffer_level),
      uptime: uptimeFromRaw(raw.metrics.uptime),
    },
    events: raw.events.map(metricEventFromRaw),
  };
}

export function dashboardStatsFromRaw(raw: RawDashboardStats): DashboardStats {
  return {
    devices: {
      total: raw.devices.total,
      online: raw.devices.online,
      offline: raw.devices.offline,
    },
    commands: {
      totalToday: raw.commands.total_today,
      pending: raw.commands.pending,
      failed: raw.commands.failed,
    },
    activity: {
      last24h: {
        commands: raw.activity.last_24h.commands,
        registrations: raw.activity.last_24h.registrations,
        deregistrations: raw.activity.last_24h.deregistrations,
      },
    },
  };
}
