import React from "react";
import ReactDOMServer from "react-dom/server";

import App from "./App";

export function render(_url: string, prefetchedState: unknown) {
  // Irrigates state into global variables inside Node/Nitro context
  if (typeof global !== "undefined") {
    (global as Record<string, unknown>).__VYZORIX_PREFETCHED_STATE__ = prefetchedState;
  }

  // Generate plain static markup
  const html = ReactDOMServer.renderToString(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );

  return { html };
}
