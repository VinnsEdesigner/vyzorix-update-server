export const VYZOR_ANALYTICS_EVENTS = {
  appLoaded: 'app_loaded',
  pageViewed: 'page_viewed',
  errorReported: 'error_reported',
  updatePushed: 'update_pushed',
  updateCancelled: 'update_cancelled',
  updatesSynced: 'updates_synced',
  deviceSelected: 'device_selected',
  commandDispatched: 'command_dispatched',
  loginSucceeded: 'login_succeeded',
  loginFailed: 'login_failed',
  consentDecision: 'consent_decision',
} as const;

export type VyzorAnalyticsEventName =
  (typeof VYZOR_ANALYTICS_EVENTS)[keyof typeof VYZOR_ANALYTICS_EVENTS];
