/**
 * REST Client Index
 * 
 * Re-exports all REST client utilities.
 */

export {
  getRESTConfig,
  buildUrl,
  apiFetch,
  apiGet,
  apiPost,
  apiPut,
  apiPatch,
  apiDelete,
  type RESTConfig,
  type FetchOptions,
  type APIErrorResponse,
  isAPIError,
} from "./rest-client";