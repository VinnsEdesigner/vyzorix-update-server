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
					// Wire DTO types come from the api-client spec schema module,
					// not the domain barrel (the root barrel is domain types).
					schemas: "@vyzorix/api-client/generated/schemas",
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
			},
		},
	},
};
