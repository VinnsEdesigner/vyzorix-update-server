import { consumeLastCapturedError } from "./lib/error-capture";
import { renderErrorPage } from "./lib/error-page";
import { getPrefetchedAuthState } from "./lib/server/cookie-reader";
import { injectStateIntoHtml } from "./lib/server/state-injector";

interface ServerEntry {
  fetch: (request: Request, env: unknown, ctx: unknown) => Promise<Response> | Response;
}

let serverEntryPromise: Promise<ServerEntry> | undefined;

const getServerEntry = (): Promise<ServerEntry> => {
  serverEntryPromise ??= import("@tanstack/react-start/server-entry").then(
    (m) => (m.default ?? m) as ServerEntry,
  );
  return serverEntryPromise;
};

// h3 swallows in-handler throws into a normal 500 Response with body
// {"unhandled":true,"message":"HTTPError"} — try/catch alone never fires for those.
const normalizeCatastrophicSsrResponse = async (response: Response): Promise<Response> => {
  if (response.status < 500) return response;
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) return response;

  const body = await response.clone().text();
  if (!body.includes('"unhandled":true') || !body.includes('"message":"HTTPError"')) {
    return response;
  }

  console.error(consumeLastCapturedError() ?? new Error(`h3 swallowed SSR error: ${body}`));
  return new Response(renderErrorPage(), {
    status: 500,
    headers: { "content-type": "text/html; charset=utf-8" },
  });
};

/**
 * Inject auth state into HTML responses
 *
 * This modifies the HTML response to include the authenticated operator state
 * from server-side cookie reading, enabling SSR hydration without flash.
 *
 * Based on Library's server.ts pattern:
 *   const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;
 *   return html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
 */
const injectAuthState = async (request: Request, response: Response): Promise<Response> => {
  // Only process HTML responses
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("text/html")) {
    return response;
  }

  try {
    // Read the HTML body
    const html = await response.text();

    // Get auth state from cookies (server-side)
    const authState = await getPrefetchedAuthState(request);

    // Inject state into HTML
    const hydratedHtml = injectStateIntoHtml(html, authState);

    // Log for debugging
    if (authState.isAuthenticated) {
      console.log(`[SSR] Injecting auth state for: ${authState.operator?.email}`);
    } else {
      console.log("[SSR] No auth state (unauthenticated)");
    }

    // Return new response with injected state
    return new Response(hydratedHtml, {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
    });
  } catch (error) {
    console.error("[SSR] Failed to inject auth state:", error);
    // Return original response if injection fails
    return response;
  }
};

export default {
  async fetch(request: Request, env: unknown, ctx: unknown) {
    try {
      const handler = await getServerEntry();
      const response = await handler.fetch(request, env, ctx);

      // Inject auth state into HTML responses for SSR hydration
      const responseWithState = await injectAuthState(request, response);

      return await normalizeCatastrophicSsrResponse(responseWithState);
    } catch (error) {
      console.error(error);
      return new Response(renderErrorPage(), {
        status: 500,
        headers: { "content-type": "text/html; charset=utf-8" },
      });
    }
  },
};
