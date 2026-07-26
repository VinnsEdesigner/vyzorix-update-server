// @vyzorix/web - Client-side entry point for SSR hydration
// This file handles client-side hydration of server-rendered HTML

import { StartClient } from "@tanstack/react-start/client";
import { startTransition } from "react";
import { hydrateRoot } from "react-dom/client";

// Hydrate the application
startTransition(() => {
  hydrateRoot(document.body, <StartClient />);
});
