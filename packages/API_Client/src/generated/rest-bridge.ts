import type { AxiosRequestConfig } from 'axios';
import { restClient } from '../vyzorServer/rest/_shared/rest-client';

// Mutator used by every orval-generated endpoint function. Delegates to the
// restClient ,
// orval generates URLs without the /v1 prefix (the OpenAPI spec has no
// basePath), so we prepend it here to match the server's route mounting.

const API_PREFIX = '/v1';

function buildUrl(url: string | undefined): string {
	if (!url) return API_PREFIX;
	if (url.startsWith('/v1')) return url;
	return `${API_PREFIX}${url.startsWith('/') ? '' : '/'}${url}`;
}

function mergeConfig(
	config: AxiosRequestConfig,
	options?: AxiosRequestConfig,
): AxiosRequestConfig {
	const merged: AxiosRequestConfig = { ...config, ...(options ?? {}) };
	merged.url = buildUrl(config.url);
	if (merged.headers === undefined) {
		merged.headers = {};
	}
	return merged;
}

/** Options accepted from generated callers. The react-query SDK passes a
 * fetch-style RequestInit ({ method, headers: HeadersInit, body }); the axios
 * SDK passes AxiosRequestConfig. The `headers` member is typed as HeadersInit
 * so generated hooks' header access (`options?.headers`) satisfies the
 * fetch-style getHeaders helper. */
export type GeneratedCallOptions = Omit<RequestInit, 'headers'> & { headers?: HeadersInit };

// Normalize fetch-style options (RequestInit) into an axios config.
function toAxiosConfig(opts: GeneratedCallOptions | undefined): AxiosRequestConfig {
	if (!opts) return {};
	const o = opts as RequestInit & { headers?: HeadersInit };
	const cfg: AxiosRequestConfig = {};
	if (o.method) cfg.method = o.method;
	if (o.body !== undefined) cfg.data = o.body;
	if (o.headers !== undefined) {
		const h = o.headers;
		cfg.headers = h instanceof Headers
			? Object.fromEntries(h.entries())
			: Array.isArray(h)
				? Object.fromEntries(h)
				: (h as Record<string, string>);
	}
	if (o.signal) cfg.signal = o.signal;
	return cfg;
}

export async function customAxios<T>(
	config: AxiosRequestConfig | string,
	options?: GeneratedCallOptions,
): Promise<T> {
	// Support both generated call shapes: orval's axios SDK passes a full
	// config object; the react-query SDK passes the URL as the first arg
	// with fetch-style options in the second.
	const optsCfg = toAxiosConfig(options);
	const cfg: AxiosRequestConfig = typeof config === 'string'
		? { ...optsCfg, url: config }
		: { ...config };
	const method = (cfg.method ?? 'GET').toUpperCase();
	const merged = mergeConfig(cfg, optsCfg);

	switch (method) {
		case 'GET':
			return restClient.get<T>(merged.url!, merged);
		case 'POST':
			return restClient.post<T>(merged.url!, merged.data, merged);
		case 'PUT':
			return restClient.put<T>(merged.url!, merged.data, merged);
		case 'PATCH':
			return restClient.patch<T>(merged.url!, merged.data, merged);
		case 'DELETE':
			return restClient.delete<T>(merged.url!, merged);
		default:
			throw new Error(`Unsupported HTTP method: ${method}`);
	}
}

export default customAxios;
