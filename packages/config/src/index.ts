// @vyzorix/config - Main exports
// Re-export all public APIs from submodules

export * from "./vite";
export * from "./auth";
export * from "./api";
export * from "./env";
export * from "./presets";
export * from "./nitro";

// Main configuration function
import { defineViteConfig } from "./vite";
import { createAuthClient } from "./auth";
import { createApiClient } from "./api";
import { loadEnv } from "./env";
import { presets, getPreset } from "./presets";
import { nitroTargets, getNitroTarget } from "./nitro";

export {
  defineViteConfig,
  createAuthClient,
  createApiClient,
  loadEnv,
  presets,
  getPreset,
  nitroTargets,
  getNitroTarget
};

// Package info
export const PACKAGE_NAME = "@vyzorix/config";
export const VERSION = "1.0.0";
export const DESCRIPTION = "Comprehensive configuration package for Vyzorix ecosystem projects";