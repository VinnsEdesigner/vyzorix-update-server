
export { graphqlClient, getApolloClient, type GraphQLConfig } from './graphql-client';

// Diagnostics normalizers — used by the web hooks' GraphQL fallback. The
// generated typed documents return wire shapes; these map them onto the domain
// DeviceInspection/TimelineResult the UI consumes.
export * from './diagnostics-types';
export { graphqlDeviceInspectionFromRaw, graphqlTimelineResultFromRaw } from './diagnostics-mappers';
