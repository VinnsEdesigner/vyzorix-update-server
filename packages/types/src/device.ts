// Device types for Vyzorix
export interface Device {
  id: string
  deviceId: string
  appVersion: string
  deviceClass: string
  firebaseInstallId?: string
  fcmToken?: string
  online: boolean
  lastSeen: number
}

export interface DeviceStatus {
  deviceId: string
  online: boolean
  lastSeen: number
  appVersion: string
  deviceClass: string
}
