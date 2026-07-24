export const DEVICE_FRAGMENT = `
  fragment Device on Device {
    id
    imei
    device_name
    model
    manufacturer
    os_version
    app_version
    status
    registered_at
    last_seen
    fcm_token_valid
    command_secret_set
  }
`;
