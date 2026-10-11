// query-key-mutator.ts — orval `override.query.queryKey` mutator.
//
// Prepends the current organization id to every generated query key so the
// unified key shape is `[organizationId, ...orvalKey]`, mirroring the hand-
// written hooks (which scope their keys by org). This is what lets generated
// `invalidateQueries` calls and the hooks' cached entries match, so automatic
// invalidation actually fires (VYZORIX_FRONTEND_ARCHITECTURE §1).
//
// The org id comes from the api-client's in-memory authContext (the source of
// truth); importing it from `@vyzorix/api-client` is safe — api-client has no
// dependency on this generated output, so there is no import cycle.
import { getCurrentOrganizationId } from "@vyzorix/api-client";

/**
 * orval `query.queryKey` mutator. orval calls it as
 * `organizationScopedQueryKey({ <queryProps> }, { url, queryOptions })` — the
 * first arg holds the operation's query params, the second the route. We build
 * the key as `[orgId?, url, ...paramValues]`.
 *
 * The org prefix is the important part: generated queries and the generated
 * `invalidateQueries` helpers both call this, so their keys match each other;
 * hand-written hooks read the same helper via the generated `getXQueryKey`s,
 * so everything shares one org-scoped key scheme.
 */
export function organizationScopedQueryKey(
  props?: Record<string, unknown>,
  meta?: { url?: string; queryOptions?: unknown },
): unknown[] {
  const url = meta?.url ?? "";
  const values = props ? Object.values(props).filter((v) => v !== undefined) : [];
  const base = [url, ...values];
  const org = getCurrentOrganizationId();
  return org ? [org, ...base] : base;
}

export default organizationScopedQueryKey;
