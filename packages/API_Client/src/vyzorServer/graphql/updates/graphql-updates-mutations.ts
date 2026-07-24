import { gql } from '@apollo/client';
import { graphqlClient } from '../_shared/graphql-client';

export const REQUEST_UPDATE = gql`
  mutation RequestUpdate($organizationId: ID!, $deviceId: ID!, $version: String!) {
    requestUpdate(organizationId: $organizationId, deviceId: $deviceId, version: $version) {
      success
      message
    }
  }
`;

export const APPLY_UPDATE = gql`
  mutation ApplyUpdate($organizationId: ID!, $deviceId: ID!, $version: String!) {
    applyUpdate(organizationId: $organizationId, deviceId: $deviceId, version: $version) {
      success
      message
    }
  }
`;

export async function mutateRequestUpdate(params: { organizationId: string; deviceId: string; version: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REQUEST_UPDATE,
    variables: params,
  });
}

export async function mutateApplyUpdate(params: { organizationId: string; deviceId: string; version: string }) {
  return graphqlClient.getClient().mutate({
    mutation: APPLY_UPDATE,
    variables: params,
  });
}
