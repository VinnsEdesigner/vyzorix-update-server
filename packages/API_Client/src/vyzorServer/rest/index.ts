// Hand-rolled REST endpoints were superseded by the orval-generated SDK
// (src/generated/*, exported from the package root as getX() accessors).
// What remains here is transport infrastructure and endpoint groups with no
// OpenAPI counterpart.

// Re-export shared transport (restClient, token/CSRF state, connectivity).
export * from "./_shared";

export * from "./health";
export * from "./oauth";
