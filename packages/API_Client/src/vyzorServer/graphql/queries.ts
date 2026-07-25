export {
  GET_INBOX_ENTRIES,
  GET_INBOX_ENTRY,
  queryInboxEntries,
  queryInboxEntry,
} from "./queries/registration-queries";

export {
  GET_DEVICES,
  GET_DEVICE,
  GET_DEVICE_TELEMETRY,
  queryDevices,
  queryDevice,
  queryDeviceTelemetry,
} from "./queries/registration-queries";

export {
  GET_DEVICES as GET_DEVICE_LIST,
  GET_DEVICE as GET_DEVICE_BY_ID,
  GET_DEVICE_COUNT,
  queryDevices as queryDeviceList,
  queryDevice,
  queryDeviceCount,
} from "./device";

export {
  GET_SETTINGS,
  GET_DEVICE_SETTINGS,
  GET_ORGANIZATION_SETTINGS,
  GET_THRESHOLDS,
  querySettings,
  queryDeviceSettings,
  queryOrganizationSettings,
  queryThresholds,
} from "./settings";

export {
  GET_UPDATES,
  queryUpdates,
} from "./updates";

export {
  GET_DEVICE_INSPECTION,
  GET_DEVICE_TIMELINE,
  queryDeviceInspection,
  queryDeviceTimeline,
} from "./diagnostics";

export {
  GET_LOGS,
  GET_LOG_ENTRY,
  queryLogs,
  queryLogEntry,
} from "./logs";

export {
  GET_PENDING_COMMANDS,
  GET_COMMAND,
  queryPendingCommands,
  queryCommand,
} from "./commands";

export {
  GET_ORGANIZATIONS,
  GET_ORGANIZATION,
  GET_MY_MEMBERSHIPS,
  GET_ORGANIZATION_MEMBERS,
  GET_ORGANIZATION_INVITATIONS,
  GET_INVITATION_BY_TOKEN,
} from "./organization";
