// @vyzorix/config/scripts/scaffolders/vscode.ts - VSCode scaffolding
import { writeFile, mkdir } from "fs/promises";
import { join } from "path";

export async function scaffoldVSCode(_target: string): Promise<void> {
  const vscodeDir = join(process.cwd(), ".vscode");
  await mkdir(vscodeDir, { recursive: true });

  const settings = {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.codeActionsOnSave": {
      "source.fixAll.eslint": "explicit",
    },
    "editor.tabSize": 2,
    "typescript.tsdk": "node_modules/typescript/lib",
    "tailwindCSS.includeLanguages": {
      "typescriptreact": "html",
    },
    "eslint.validate": [
      "javascript",
      "javascriptreact",
      "typescript",
      "typescriptreact",
    ],
    "prettier.requireConfig": true,
    "files.eol": "\n",
    "files.insertFinalNewline": true,
    "search.exclude": {
      "**/node_modules": true,
      "**/dist": true,
      "**/build": true,
      "**/.next": true,
      "**/coverage": true,
    },
  };

  const extensions = {
    recommendations: [
      "dbaeumer.vscode-eslint",
      "esbenp.prettier-vscode",
      "bradlc.vscode-tailwindcss",
      "ms-vscode.vscode-typescript-next",
      "eamodio.gitlens",
    ],
  };

  await writeFile(
    join(vscodeDir, "settings.json"),
    JSON.stringify(settings, null, 2)
  );
  await writeFile(
    join(vscodeDir, "extensions.json"),
    JSON.stringify(extensions, null, 2)
  );
}