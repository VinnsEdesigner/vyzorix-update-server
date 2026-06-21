#!/usr/bin/env python3
"""
Vyzorix Agent Task Runner
========================
Delegates coding tasks to SiliconFlow AI and executes them in the codebase.

Usage:
    export SILICONFLOW_API_KEY="your-key"
    python3 agent_task_runner.py "Your task description here"
"""

import os
import sys
import subprocess
from pathlib import Path

# Project root is parent of AGENTS/
PROJECT_ROOT = Path(__file__).parent.parent
AGENTS_DIR = Path(__file__).parent

def run_command(cmd: str, cwd: Path = PROJECT_ROOT) -> tuple[str, str, int]:
    """Run a shell command and return stdout, stderr, exit code."""
    result = subprocess.run(
        cmd,
        shell=True,
        cwd=cwd,
        capture_output=True,
        text=True
    )
    return result.stdout, result.stderr, result.returncode

def read_file(path: str) -> str:
    """Read and return file contents."""
    full_path = PROJECT_ROOT / path
    try:
        with open(full_path, 'r') as f:
            return f.read()
    except FileNotFoundError:
        return f"File not found: {path}"
    except Exception as e:
        return f"Error reading {path}: {e}"

def get_file_list(pattern: str = "**/*.go", max_depth: int = 5) -> list:
    """Get list of files matching pattern."""
    files = []
    for f in PROJECT_ROOT.glob(pattern):
        if f.is_file():
            rel_path = f.relative_to(PROJECT_ROOT)
            depth = len(rel_path.parts)
            if depth <= max_depth:
                files.append(str(rel_path))
    return files

def get_project_context() -> dict:
    """Gather project context for the agent."""
    context = {
        "project_root": str(PROJECT_ROOT),
        "structure": {},
        "git_status": None,
        "go_mod": None,
    }
    
    # Get directory structure (top 3 levels)
    for d in PROJECT_ROOT.rglob("*"):
        if d.is_dir() and "/.git" not in str(d):
            rel = d.relative_to(PROJECT_ROOT)
            parts = rel.parts
            if len(parts) <= 4:
                context["structure"][str(rel)] = [f.name for f in d.iterdir() if f.is_file()][:10]
    
    # Git status
    stdout, _, _ = run_command("git status --short")
    context["git_status"] = stdout if stdout else "Clean"
    
    # Go module
    go_mod = read_file("apps/api/go.mod")
    context["go_mod"] = go_mod[:500] if len(go_mod) > 500 else go_mod
    
    return context

def delegate_to_agent(task: str, context: dict) -> str:
    """Send task to SiliconFlow agent."""
    sys.path.insert(0, str(AGENTS_DIR))
    from siliconflow_agent import coding_agent
    
    # Build context prompt
    context_prompt = f"""
Project Context:
- Root: {context['project_root']}
- Git Status: {context['git_status']}
- Directory Structure: {list(context['structure'].keys())[:20]}

Task: {task}

Important:
1. Read files using absolute paths from project root
2. Use subprocess to run commands like: go test ./..., go build ./...
3. Make changes using file_editor tool
4. Always work from: {context['project_root']}
"""
    
    agent = coding_agent()
    response = agent(context_prompt)
    return response

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 agent_task_runner.py 'Your task description'")
        print("\nExample:")
        print("  python3 agent_task_runner.py 'Review the auth handlers for security issues'")
        sys.exit(1)
    
    task = " ".join(sys.argv[1:])
    
    print(f"🤖 Delegating task to SiliconFlow AI...")
    print(f"   Task: {task[:100]}...")
    print()
    
    # Get context
    context = get_project_context()
    
    # Delegate
    response = delegate_to_agent(task, context)
    
    print("\n" + "="*60)
    print("AGENT RESPONSE:")
    print("="*60)
    print(response)

if __name__ == "__main__":
    main()
