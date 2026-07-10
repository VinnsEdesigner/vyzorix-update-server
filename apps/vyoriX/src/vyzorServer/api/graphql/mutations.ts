/**
 * GraphQL Mutations
 * 
 * Barrel export for all GraphQL mutations.
 * Re-exports from subdirectory index files.
 */

// Registration mutations
export {
  SUBMIT_REGISTRATION_REQUEST,
  ACKNOWLEDGE_REQUEST,
  REGISTER_DEVICE,
  DISMISS_INBOX_ENTRY,
  CONFIRM_REGISTRATION,
  DEREGISTER_DEVICE,
} from "./mutations/registration-mutations";
