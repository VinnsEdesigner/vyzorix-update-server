import { graphqlClient } from '../_shared/graphql-client';

export const GET_DIAGNOSTICS = `
  query GetDiagnostics($organizationId: ID!, $deviceId: ID!) {
    diagnostics(organizationId: $organizationId, deviceId: $deviceId) {
      device_id
      connection_status {
        web_socket_status
        connected_at
        protocol
        client_ip
      }
      last_seen
      fcm_token_valid
      command_secret_set
    }
  }
`;

export async function queryDiagnostics(params: { organizationId: string; deviceId: string }) {
  return graphqlClient.getClient().query({
    query: GET_DIAGNOSTICS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
