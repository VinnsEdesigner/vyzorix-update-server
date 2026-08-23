import type { AxiosRequestConfig } from 'axios';
import { restClient } from '../vyzorServer/rest/_shared/rest-client';

// Mutator used by every orval-generated endpoint function. Delegates to the
// enterprise restClient (circuit breaker, HMAC signing, token refresh, retry
// with backoff, idempotency keys, offline queue, request batching) so the
// generated functions inherit all transport infrastructure without duplicating
// it. Returns Promise<T> (the response data) directly — callers don't need to
// unwrap .data.
//
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

export async function customAxios<T>(
	config: AxiosRequestConfig,
	options?: AxiosRequestConfig,
): Promise<T> {
	const method = (config.method ?? 'GET').toUpperCase();
	const merged = mergeConfig(config, options);

	switch (method) {
		case 'GET':
			return restClient.get<T>(merged.url!, merged);
		case 'POST':
			return restClient.post<T>(merged.url!, config.data, merged);
		case 'PUT':
			return restClient.put<T>(merged.url!, config.data, merged);
		case 'PATCH':
			return restClient.patch<T>(merged.url!, config.data, merged);
		case 'DELETE':
			return restClient.delete<T>(merged.url!, merged);
		default:
			throw new Error(`Unsupported HTTP method: ${method}`);
	}
}

export default customAxios;
