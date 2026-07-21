


 Vyzorix Unified Tamagui Architecture Specification


# Vyzorix Universal UI Specification
 Complete Monorepo Migration Strategy & Single-Source Tamagui Token Design


 **Author:** Senior UI Engineer

 **Target Runtimes: Android Runtime & V8 Web Dom Engine

 **Workspace Engine: pnpm + Turborepo Monorepo


## 1. The Vyzorix Cross-Platform Paradigm Pivot


 To eliminate structural drift, presentation logic discrepancies, and token divergence, Vyzorix completely replaces string-based Tailwind configurations with a single type-safe compilation core running Tamagui. This architectural pivot moves styling from separate layout abstractions (web DOM class tracking vs mobile native stylesheets) into a shared, unified packages pipeline.


 The shared internal design token engine maps interfaces to flat atomic CSS variables when compiling for browser distributions, and translates layouts to multi-threaded native view arrays when assembling Android binaries.


 **Unified Workspace Directory Topology

/ (Workspace Root)
├── package.json
├── turbo.json
├── pnpm-workspace.yaml
├── packages/
│ ├── api-client/ ➔ Shared Data Protocol Core (GraphQL & Axios Instances)
│ └── ui/ ➔ NEW: Shared Sovereign Tamagui Design System System Core Package
│ ├── package.json
│ └── src/
│ ├── tamagui.config.ts # Core Token and Theme Configuration Object Map
│ └── components/ # Multi-Runtime Pure Shareable Layout Elements
└── apps/
 ├── web/ ➔ Vite + React 18 Engine Workspace (Consumes @vyzorix/ui)
 └── mobile/ ➔ Expo Mobile Application Layer (Consumes @vyzorix/ui)
 └── app.json ➔ Core Android OS Configuration Manifest Mapping


## 2. Sovereign Token Infrastructure Configuration


 Every variable, design token, color matrix, layout space metric, and animation timing is isolated within packages/ui/src/tamagui.config.ts. This serves as the master source-of-truth.

 ```
import { createAnimations } from '@tamagui/animations-react-native';
import { createTamagui, createTokens } from 'tamagui';

const animations = createAnimations({
 bouncy: {
 type: 'spring',
 damping: 11,
 mass: 0.9,
 stiffness: 110,
 },
 lazy: {
 type: 'spring',
 damping: 20,
 stiffness: 65,
 },
});

const tokens = createTokens({
 size: { 0: 0, 1: 4, 2: 8, 3: 12, 4: 16, true: 16 },
 space: { 0: 0, 1: 4, 2: 8, 3: 12, 4: 16, true: 16 },
 radius: { 0: 0, 1: 4, 2: 8, true: 8 },
 color: {
 brandDark: '#020617',
 brandPurple: '#a855f7',
 brandGreen: '#34d399',
 textWhite: '#f8fafc',
 },
});

export const config = createTamagui({
 animations,
 defaultTheme: 'dark',
 shouldAddPrefersColorThemes: false,
 themeClassNameOnRoot: false,
 tokens,
 media: {
 xs: { maxWidth: 660 },
 sm: { maxWidth: 800 },
 md: { maxWidth: 1020 },
 gtMd: { minWidth: 1021 },
 },
 themes: {
 dark: {
 background: tokens.color.brandDark,
 color: tokens.color.textWhite,
 primary: tokens.color.brandPurple,
 success: tokens.color.brandGreen,
 }
 }
});

export type AppConfig = typeof config;
declare module 'tamagui' {
 interface TamaguiCustomConfig extends AppConfig {}
}
export default config;
```


## 3. The Sovereign Shared Components Factory


 Inside packages/ui/src/components/EngineCard.tsx, components are written using layout stacks that map correctly across different runtimes. Web targets receive mouse hovers, responsive breakpoints scale layout bounds, and mobile platforms map directly to hardware targets.

 ```
import { YStack, XStack, Text, Button } from 'tamagui';

export function EngineCard() {
 return (
 <YStack
 backgroundColor="$background"
 padding="$4"
 borderRadius="$radius.true"
 borderWidth={1}
 borderColor="rgba(168, 85, 247, 0.25)"
 width="100%"
 $gtMd={{ width: '50%' }}
 >
 <XStack justifyContent="space-between" alignItems="center">
 <Text color="$color" fontSize="$5" fontWeight="bold">
 Vyzorix Worker Node
 </Text>
 <Text color="$success" fontSize="$2" fontStyle="mono">
 ● ONLINE
 </Text>
 </XStack>

 <Button
 marginTop="$4"
 backgroundColor="$primary"
 animation="bouncy"
 hoverStyle={{ backgroundColor: '#c084fc', scale: 1.02 }}
 pressStyle={{ backgroundColor: '#7e22ce', scale: 0.98 }}
 >
 <Text color="#fff" fontWeight="bold">Sync Operational Clusters</Text>
 </Button>
 </YStack>
 );
}
```


## 4. Web Runtime Compiler Engineering


 The web environment must hook the Tamagui optimizing static evaluator directly into Vite. This forces string-less props to unpack cleanly into un-bloated, highly cached atomic CSS entries during production deployments.


### 4.1 Vite Build Configuration: apps/web/vite.config.ts
 ```
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { tamaguiPlugin } from '@tamagui/vite-plugin';

export default defineConfig({
 plugins: [
 react(),
 tamaguiPlugin({
 config: '../../packages/ui/src/tamagui.config.ts',
 components: ['tamagui'],
 }),
 ],
});
```


## 5. Native Mobile Runtime Compiler Engineering


 The native environment forces the Metro bundler to read and execute transformations. The compiler hooks directly into the compilation layers, converting components into low-level view code.


### 5.1 Metro Config Pipeline mapping: apps/mobile/metro.config.js
 ```
const { getDefaultConfig } = require('expo/metro-config');
const { withTamagui } = require('@tamagui/metro-plugin');
const path = require('path');

const projectRoot = __dirname;
const workspaceRoot = path.resolve(projectRoot, '../..');

let config = getDefaultConfig(projectRoot);

// Bind the package workspace path tracking monitors
config.watchFolders = [workspaceRoot];
config.resolver.nodeModulesPaths = [
 path.resolve(projectRoot, 'node_modules'),
 path.resolve(workspaceRoot, 'node_modules'),
];
config.resolver.disableHierarchicalLookup = true;

module.exports = withTamagui(config, {
 config: '../../packages/ui/src/tamagui.config.ts',
 components: ['tamagui'],
 outputCSS: null,
});
```


## 6. Complete Structural Configuration Packages


 The package.json file for the internal UI framework must export entry paths so that pnpm workspace links can resolve types and modules headlessly across applications.


### 6.1 Shared UI Definition: packages/ui/package.json
 ```
{
 "name": "@vyzorix/ui",
 "version": "1.0.0",
 "main": "./src/index.ts",
 "types": "./src/index.ts",
 "peerDependencies": {
 "react": "18.x",
 "react-native": "*"
 },
 "dependencies": {
 "tamagui": "^1.90.0",
 "@tamagui/animations-react-native": "^1.90.0",
 "@tamagui/config": "^1.90.0"
 }
}
```


## 7. Platform Dual-Execution Matrix Summary


 Feature Paradigm
 Web Runtime Resolution (Vite Stack)
 Mobile Runtime Resolution (Metro Native)


 **1. Core Element Output
 Transforms components into atomic <div> wrappers.
 Compiles directly to native android.view.ViewGroup classes.


 **2. Layout Mechanics
 Generates dynamic CSS style rules and appends standard flex mappings.
 Processes layout metrics cleanly on the GPU via the Yoga layout engine.


 **3. Interactivity Tracking
 Attaches standard :hover mouse and focus classes.
 Safely skips mouse rules, mapping actions directly to native touch listeners.


 **4. Animation Execution
 Maps transitions to lightweight CSS keyframe strings.
 Triggers frame-accurate, hardware-accelerated physics animations.


 ️ Network Cleartext & API Sandbox Verification


 As this UI system maps directly to your active Go Gin server instances, verify your backend connection profiles. Maintain secure cross-origin settings inside the Go Gin router definitions to handle both local dev sandboxes and live web routes without causing security blocks.


