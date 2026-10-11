const spec = require("fs").existsSync("apps/api/swag/openapi3.json")
        ? "apps/api/swag/openapi3.json"
        : "../../apps/api/swag/openapi3.json";

module.exports = {
        // Axios function objects (used by orchestration hooks + restClient transport)
        sdk: {
                input: spec,
                output: {
                        target: "packages/API_Client/src/generated",
                        mode: "tags-split",
                        client: "axios",
                        override: {
                                mutator: {
                                        path: "packages/API_Client/src/generated/rest-bridge.ts",
                                        name: "customAxios",
                                },
                                npmPackage: "@vyzorix/api-client",
                        },
                },
        },
        // TanStack Query hooks (replaces hand-written hooks in apps/VyzoriX_web/src/hooks/)
        // Output goes to the web app directly — the organize_query_hooks.py script
        // renames the generic tag files to explicit Vyzorix-domain filenames.
        vyzorixRQ: {
                input: spec,
                output: {
                        target: "apps/VyzoriX_web/src/generated-rq",
                        // orval detects the React Query major from the package.json it
                        // reads; the workspace root declares none, so point it at the web
                        // app's so it emits React Query v5 signatures (4-arg onSuccess).
                        packageJson: "apps/VyzoriX_web/package.json",
                        // Wire DTO types come from the api-client spec schema module, not a
                        // the domain barrel (the root barrel is domain types).
                        schemas: "packages/API_Client/src/generated/schemas",
                        mode: "tags-split",
                        client: "react-query",
                        override: {
                                mutator: {
                                        path: "packages/API_Client/src/generated/rest-bridge.ts",
                                        name: "customAxios",
                                },
                                // The bridge returns the response body (not the AxiosResponse
                                // envelope), so generated hooks are typed against the body shape.
                                fetch: {
                                        includeHttpResponseReturnType: false,
                                },
                                // Automatic cache invalidation: each mutating operation
                                // invalidates the query operations in its tag, so mutations
                                // refetch related queries without hand-written
                                // invalidateQueries() calls (VYZORIX_FRONTEND_ARCHITECTURE §1).
                                query: {
                                        version: 5,
                                        useInvalidate: true,
                                        shouldExportQueryKey: true,
                                        // Prefix every generated query key with the active
                                        // organization id, so generated invalidation matches
                                        // the org-scoped keys the hand-written hooks use.
                                        queryKey: {
                                                path: "apps/VyzoriX_web/src/generated-rq/query-key-mutator.ts",
                                                name: "organizationScopedQueryKey",
                                                default: true,
                                        },
                                        mutationInvalidates: [
				{ onMutations: ['deleteAdminApiKeysKeyId', 'deleteAdminClientsClientId', 'patchAdminClientsClientId', 'postAdminClientsClientIdRotateKey', 'postAuthAdminOperators', 'deleteAuthAdminOperatorsId', 'patchAuthAdminOperatorsId'], invalidates: [{ query: 'getAdminApiKeys' }, { query: 'getAdminApiKeysStats' }, { query: 'getAdminClients' }, { query: 'getAdminSupportBundle' }, { query: 'getAuthAdminOperators' }] },
				{ onMutations: ['postAlertsRules', 'deleteAlertsRulesId', 'patchAlertsRulesId', 'postAlertsRulesIdEvaluate'], invalidates: [{ query: 'getAlertsRules' }] },
				{ onMutations: ['postAnnotations', 'deleteAnnotationsId', 'patchAnnotationsId'], invalidates: [{ query: 'getAnnotations' }] },
				{ onMutations: ['postAuthApiKeys', 'deleteAuthApiKeysKeyId', 'patchAuthApiKeysKeyId', 'postAuthApiKeysKeyIdRotate'], invalidates: [{ query: 'getAuthApiKeys' }] },
				{ onMutations: ['postAuthAdminLockoutUnlockOperatorId', 'postAuthCancelVerification', 'postAuthForgotPassword', 'postAuthLogin', 'postAuthLoginTokens', 'postAuthLogout', 'postAuthOauthGithub', 'postAuthOauthGoogle', 'postAuthOrganizationsSelect', 'postAuthRefresh', 'postAuthRegister', 'postAuthResendPasswordReset', 'postAuthResendVerification', 'postAuthResetPassword', 'postAuthVerifyEmail'], invalidates: [{ query: 'getAuthLockoutStatus' }, { query: 'getAuthMe' }, { query: 'getAuthOrganizations' }, { query: 'getAuthPollVerification' }, { query: 'getAuthResendVerification' }, { query: 'getAuthVerifyEmail' }] },
				{ onMutations: ['postChannelsSubscribe', 'postChannelsUnsubscribe'], invalidates: [{ query: 'getChannelsStatus' }] },
				{ onMutations: ['postAuthClientCredentials', 'deleteAuthClientCredentialsClientId', 'patchAuthClientCredentialsClientId', 'postAuthClientCredentialsClientIdRotateSecret'], invalidates: [{ query: 'getAuthClientCredentials' }] },
				{ onMutations: ['postNotificationsContactPoints', 'deleteNotificationsContactPointsId', 'patchNotificationsContactPointsId', 'postNotificationsContactPointsIdTest'], invalidates: [{ query: 'getNotificationsContactPoints' }] },
				{ onMutations: ['postDeviceConfirm', 'deleteDeviceImei', 'postDeviceImeiDisconnect', 'patchDeviceImeiFcmToken', 'deleteDevicesImei', 'patchDevicesImeiSettings', 'patchDevicesImeiSettingsThresholds', 'putDevicesImeiTags', 'postDevicesImeiTagsTag', 'deleteDevicesImeiTagsTag', 'postOrganizationsIdDevicesImeiTransfer'], invalidates: [{ query: 'getDashboardDevices' }, { query: 'getDashboardDevicesOperator' }, { query: 'getDashboardEventsRecent' }, { query: 'getDeviceCount' }, { query: 'getDevices' }] },
				{ onMutations: ['postDeviceInbox', 'patchDeviceInboxImei', 'postDeviceInboxImeiAck', 'postDeviceInboxImeiResend'], invalidates: [{ query: 'getDeviceInbox' }] },
				{ onMutations: ['postInvitations', 'deleteInvitationsId', 'postInviteTokenAccept', 'postInviteTokenReject'], invalidates: [{ query: 'getInvitations' }, { query: 'getMeInvitations' }] },
				{ onMutations: ['postAuthMfaDisable', 'postAuthMfaEnable', 'postAuthMfaEnroll', 'postAuthMfaRegenerateBackupCodes', 'postAuthMfaVerify', 'postAuthMfaVerifyBackup', 'postAuthMfaVerifySetup'], invalidates: [{ query: 'getAuthMfaStatus' }] },
				{ onMutations: ['postOrganizations', 'deleteOrganizationsId', 'patchOrganizationsId', 'patchOrganizationsIdSettings', 'patchOrganizationsIdSettingsThresholds'], invalidates: [{ query: 'getOrganizations' }] },
				{ onMutations: ['postServiceAccounts', 'deleteServiceAccountsId', 'postServiceAccountsIdRotate', 'postServiceAccountsIdTokens', 'deleteServiceAccountsIdTokensToken'], invalidates: [{ query: 'getServiceAccounts' }] },
				{ onMutations: ['deleteAuthSessions', 'postAuthSessionsRevokeAll', 'deleteAuthSessionsId'], invalidates: [{ query: 'getAuthSessions' }, { query: 'getAuthSessionsConcurrent' }] },
				{ onMutations: ['patchAuthMe', 'patchMeNotifications', 'postMeNotificationsWebhookRotate', 'postMeNotificationsWebhookTest', 'patchMePreferences', 'patchMeSettings', 'patchMeThresholds'], invalidates: [{ query: 'getMeNotifications' }, { query: 'getMePreferences' }, { query: 'getMeSettings' }, { query: 'getMeThresholds' }] },
				{ onMutations: ['deleteTelemetryCleanup'], invalidates: [{ query: 'getTelemetryHistory' }, { query: 'getTelemetryHistoryExport' }] },
				{ onMutations: ['postDownloadProgress'], invalidates: [{ query: 'getChangelog' }, { query: 'getCheckUpdate' }, { query: 'getVersion' }] },
				                                { onMutations: ['postUpdatesDeviceStatus', 'postUpdatesHistoryPushIdCancel', 'postUpdatesPush', 'postUpdatesSync'], invalidates: [{ query: 'getUpdatesChangelog' }, { query: 'getUpdatesExport' }, { query: 'getUpdatesHistory' }, { query: 'getUpdatesStatus' }, { query: 'getUpdatesSyncStatus' }, { query: 'getUpdatesVersions' }] },
                                        ],
                                },
                        },
                },
        },
};
