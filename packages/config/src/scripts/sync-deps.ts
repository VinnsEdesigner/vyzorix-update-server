// @vyzorix/config/scripts/sync-deps.ts - Sync dependencies across workspace
import { readFile, writeFile } from "fs/promises";
import { join, resolve } from "path";
import { readdir } from "fs/promises";

interface PackageJson {
  name: string;
  version: string;
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
  workspaces?: string[];
}

async function findWorkspacePackages(rootDir: string): Promise<string[]> {
  const packages: string[] = [];
  
  try {
    const entries = await readdir(rootDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const pkgPath = join(rootDir, entry.name, "package.json");
        try {
          const content = await readFile(pkgPath, "utf-8");
          const pkg = JSON.parse(content) as PackageJson;
          if (pkg.name) {
            packages.push(pkgPath);
          }
        } catch (_e) {
          // Not a package, skip
        }
      }
    }
  } catch (_e) {
    // Error reading workspace
  }
  
  return packages;
}

async function syncDependency(
  workspacePackages: string[],
  depName: string,
  depVersion: string
): Promise<void> {
  for (const pkgPath of workspacePackages) {
    try {
      const content = await readFile(pkgPath, "utf-8");
      const pkg = JSON.parse(content) as PackageJson;
      
      let updated = false;
      
      // Update dependencies
      if (pkg.dependencies && pkg.dependencies[depName]) {
        pkg.dependencies[depName] = depVersion;
        updated = true;
      }
      
      // Update devDependencies
      if (pkg.devDependencies && pkg.devDependencies[depName]) {
        pkg.devDependencies[depName] = depVersion;
        updated = true;
      }
      
      // Update peerDependencies
      if (pkg.peerDependencies && pkg.peerDependencies[depName]) {
        pkg.peerDependencies[depName] = depVersion;
        updated = true;
      }
      
      if (updated) {
        await writeFile(pkgPath, JSON.stringify(pkg, null, 2) + "\n");
        console.log(`  ✅ Updated ${pkg.name} (${depName}@${depVersion})`);
      }
    } catch (e) {
      console.error(`  ⚠️  Failed to update ${pkgPath}:`, e);
    }
  }
}

async function main() {
  console.log("🔄 Syncing dependencies across workspace...\n");
  
  const rootDir = process.cwd();
  const packageJsonPath = join(rootDir, "package.json");
  
  try {
    // Read root package.json
    const rootContent = await readFile(packageJsonPath, "utf-8");
    const rootPkg = JSON.parse(rootContent) as PackageJson;
    
    // Find workspace packages
    const workspaceDirs = [rootDir];
    
    if (rootPkg.workspaces) {
      for (const workspace of rootPkg.workspaces) {
        if (workspace.includes("*")) {
          // Glob pattern, resolve directories
          const resolved = resolve(rootDir, workspace.replace("/*", ""));
          const dirs = await readdir(resolved);
          for (const dir of dirs) {
            workspaceDirs.push(join(resolved, dir));
          }
        } else {
          workspaceDirs.push(resolve(rootDir, workspace));
        }
      }
    }
    
    // Collect all packages
    const allPackages: string[] = [];
    for (const dir of workspaceDirs) {
      const packages = await findWorkspacePackages(dir);
      allPackages.push(...packages);
    }
    
    console.log(`📦 Found ${allPackages.length} packages\n`);
    
    // Sync dependencies from root
    const depsToSync = {
      ...rootPkg.dependencies,
      ...rootPkg.devDependencies,
    };
    
    for (const [name, version] of Object.entries(depsToSync)) {
      console.log(`\n🔗 Syncing ${name}@${version}...`);
      await syncDependency(allPackages, name, version);
    }
    
    console.log("\n✅ Dependency sync complete!");
  } catch (error) {
    console.error("❌ Sync failed:", error);
    process.exit(1);
  }
}

main();