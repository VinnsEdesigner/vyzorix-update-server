export * from "./_shared";

// Re-export queries and mutations explicitly to avoid conflicts
export { 
  GET_ORGANIZATIONS,
  GET_ORGANIZATION,
  GET_MY_MEMBERSHIPS,
  GET_ORGANIZATION_MEMBERS,
  GET_ORGANIZATION_INVITATIONS,
  GET_INVITATION_BY_TOKEN,
} from "./organization";

export { GET_DEVICES } from "./device";
export { GET_LOGS } from "./logs";
export { GET_UPDATES } from "./updates";
export { GET_DEVICE_INSPECTION, GET_DEVICE_TIMELINE } from "./diagnostics";
export { SEND_COMMAND } from "./commands";
