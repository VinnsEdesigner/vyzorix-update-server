import { StrictMode } from "react";
import { createRoot, hydrateRoot } from "react-dom/client";

import App from "./App.tsx";
import "./index.css";
import { ARCHITECTURE_CONFIG } from "./lib/config";

const container = document.getElementById("root")!;

if (ARCHITECTURE_CONFIG.MODE === "SSR" && container.children.length > 0) {
  hydrateRoot(
    container,
    <StrictMode>
      <App />
    </StrictMode>,
  );
} else {
  createRoot(container).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}
