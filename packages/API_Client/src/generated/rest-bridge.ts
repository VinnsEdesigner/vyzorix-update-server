import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import axios from 'axios';
import { getRESTConfig } from '../config';

// Axios instance seeded from the shared REST client configuration. All
// generated endpoint functions call this so they inherit auth headers
// (X-Organization-ID, session, CSRF) and org scoping without duplication.
const restClientInstance: AxiosInstance = axios.create({
	baseURL: getRESTConfig().baseURL,
});

export function customAxios<T>(
	config: AxiosRequestConfig,
	options?: AxiosRequestConfig,
): Promise<AxiosResponse<T>> {
	const merged = { ...config, ...(options ?? {}) };
	if (merged.headers === undefined) {
		merged.headers = {};
	}
	return restClientInstance.request(merged);
}

export default customAxios;
