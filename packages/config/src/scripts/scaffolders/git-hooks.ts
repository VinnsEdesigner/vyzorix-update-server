// @vyzorix/config/scripts/scaffolders/git-hooks.ts - Git hooks scaffolding
import { writeFile, mkdir } from "fs/promises";
import { join } from "path";

export async function scaffoldGitHooks(_target: string): Promise<void> {
  const huskyDir = join(process.cwd(), ".husky");
  await mkdir(huskyDir, { recursive: true });

  const preCommit = `#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

pnpm lint-staged
`;

  const commitMsg = `#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

npx --no -- commitlint --edit "$1"
`;

  await writeFile(join(huskyDir, "pre-commit"), preCommit);
  await writeFile(join(huskyDir, "commit-msg"), commitMsg);
}