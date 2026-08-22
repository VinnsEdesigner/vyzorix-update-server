const spec = require("fs").existsSync("apps/api/swag/openapi3.json")
	? "apps/api/swag/openapi3.json"
	: "../../apps/api/swag/openapi3.json";

module.exports = {
	sdk: {
		input: spec,
		output: {
			target: "packages/API_Client/src/generated/orval-sdk.ts",
			mode: "single",
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
};
