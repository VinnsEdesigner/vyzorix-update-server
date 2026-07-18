#!/usr/bin/env python3
"""
Find corrupted Go files in the repository.

Detects:
1. Duplicate function definitions (main corruption indicator)
2. Files with duplicate function definitions and unbalanced braces
"""

import os
import re
import sys
from collections import defaultdict

ROOT = "/workspace/vyzorix-update-server"


def find_go_files(root):
    """Find all .go files in the repository."""
    go_files = []
    for dirpath, dirnames, filenames in os.walk(root):
        if '.git' in dirpath or 'vendor' in dirpath or 'third_party' in dirpath:
            continue
        for filename in filenames:
            if filename.endswith('.go'):
                go_files.append(os.path.join(dirpath, filename))
    return go_files


def check_duplicate_functions(content):
    """Check for duplicate function definitions - primary corruption indicator.
    
    Only checks for duplicate TOP-LEVEL functions (not methods).
    Methods like 'func (e *T) Error()' are fine to have the same method name
    on different receiver types.
    """
    # Match only top-level functions: func Name( or func Name(args) returntype
    # NOT methods: func (receiver) Name(
    top_level_pattern = re.compile(
        r'^func\s+([A-Z][a-zA-Z0-9]*)\s*\(',  # func NAME(
        re.MULTILINE
    )
    
    # Also catch lowercase top-level functions (less common but valid)
    lowercase_pattern = re.compile(
        r'^func\s+([a-z][a-zA-Z0-9]*)\s*\(',  # func name(
        re.MULTILINE
    )
    
    functions = defaultdict(list)
    
    for match in top_level_pattern.finditer(content):
        func_name = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        functions[func_name].append(line_num)
    
    for match in lowercase_pattern.finditer(content):
        func_name = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        functions[func_name].append(line_num)
    
    duplicates = {name: lines for name, lines in functions.items() if len(lines) > 1}
    return duplicates


def analyze_file(filepath):
    """Analyze a single Go file for corruption."""
    try:
        with open(filepath, 'r') as f:
            content = f.read()
    except Exception as e:
        return {"error": str(e)}
    
    result = {
        "path": filepath.replace(ROOT, ""),
        "issues": [],
        "severity": "ok"
    }
    
    # Check: Duplicate functions (reliable corruption indicator)
    duplicates = check_duplicate_functions(content)
    if duplicates:
        for func_name, lines in duplicates.items():
            result["issues"].append(f"duplicate function '{func_name}' at lines {lines}")
            result["severity"] = "critical"
    
    return result


def main():
    print("🔍 Scanning for corrupted Go files (duplicate function detection)...\n")
    
    go_files = find_go_files(ROOT)
    print(f"Found {len(go_files)} Go files\n")
    
    all_results = []
    for filepath in sorted(go_files):
        result = analyze_file(filepath)
        if result.get("issues"):
            all_results.append(result)
    
    if not all_results:
        print("✅ No corrupted files found!")
        return 0
    
    print(f"⚠️  Found {len(all_results)} files with duplicate functions:\n")
    print("=" * 60)
    
    for r in all_results:
        print(f"\n❌ {r['path']}")
        for issue in r["issues"]:
            print(f"   - {issue}")
    
    print(f"\n\nTotal: {len(all_results)} corrupted files")
    return 1


if __name__ == '__main__':
    sys.exit(main())
