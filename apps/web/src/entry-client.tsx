// @vyzorix/web - Client-side entry point for SSR hydration
// This file handles client-side hydration of server-rendered HTML

import { HydratedRouter } from "@tanstack/react-start/client";
import { startTransition } from "react";
import { hydrateRoot } from "react-dom/client";

import { getRouter } from "./router";

// Get the router instance
const router = getRouter();

// Hydrate the application
startTransition(() => {
  hydrateRoot(document.body, <HydratedRouter router={router} />);
});
