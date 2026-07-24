export const COMMAND_LIST_FRAGMENT =  `
  fragment CommandList on Command {
    id
    dispatchId
    type
    deviceImei
    status
    createdAt
  }
`;

export const COMMAND_FRAGMENT =  `
  fragment Command on Command {
    id
    dispatchId
    type
    deviceImei
    status
    params
    result {
      success
      message
      output
      error
    }
    createdAt
    sentAt
    acknowledgedAt
    completedAt
    error
  }
`;
