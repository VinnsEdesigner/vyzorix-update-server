const { getDefaultConfig } = require('expo/metro-config');
const { withTamagui } = require('@tamagui/metro-plugin');
const path = require('path');

const projectRoot = __dirname;
const workspaceRoot = path.resolve(projectRoot, '../..');

let config = getDefaultConfig(projectRoot);

// Force Metro to track all workspace files in the root monorepo directory
config.watchFolders = [workspaceRoot];

// Force module lookup to correctly fall back to root node_modules paths for pnpm symlinks
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(workspaceRoot, 'node_modules'),
];

// Block nested symbolic lookups from throwing duplication exceptions
config.resolver.disableHierarchicalLookup = true;

module.exports = withTamagui(config, {
  config: '../../packages/ui/src/tamagui.config.ts',
  components: ['tamagui'],
  outputCSS: null,
});
