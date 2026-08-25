package vyzorix

// DeviceThresholds defines alert thresholds for a device's risk,
// thermal, and buffer metrics. Warn must be less than Crit for
// each metric. Values are 0-100.
#DeviceThresholds: {
        riskWarn:    int | *0
        riskCrit:    int | *0
        thermalWarn: int | *0
        thermalCrit: int | *0
        bufferWarn:  int | *0
        bufferCrit:  int | *0
}
