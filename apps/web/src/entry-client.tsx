// @vyzorix/web - Client-side entry point for SSR hydration
// This file handles client-side hydration of server-rendered HTML

import { startTransition } from "react";
import { hydrateRoot } from "react-dom/client";
import { StartClient } from "@tanstack/react-start/client";

// Hydrate the application
startTransition(() => {
  hydrateRoot(document.body, <StartClient />);
});
