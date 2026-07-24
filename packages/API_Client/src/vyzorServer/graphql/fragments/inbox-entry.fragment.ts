export const INBOX_ENTRY_FRAGMENT = `
  fragment InboxEntry on InboxEntry {
    id
    imei
    device_name
    model
    manufacturer
    os_version
    app_version
    firmware
    security_patch
    build_id
    fcm_token
    firebase_install_id
    status
    received_at
    updated_at
    acknowledged_at
    approved_at
    rejected_at
    notes
  }
`;
