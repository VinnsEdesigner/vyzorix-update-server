#!/usr/bin/env python3
"""
Extracts all BUG-## references from Go source code comments in the API.
Usage: python3 extract_bug_fixes.py
"""

import os
import re
from collections import defaultdict
from pathlib import Path

API_PATH = Path("/workspace/project/vyzorix-update-server/apps/api")

# Pattern matches: BUG-06 Fix:, BUG-06 fix:, BUG-06: etc.
BUG_PATTERN = re.compile(r'BUG-(\d+)', re.IGNORECASE)

def scan_file(filepath: Path) -> list[tuple[str, str]]:
    """Return list of (bug_num, line_content) found in file."""
    results = []
    try:
        with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
            for line_no, line in enumerate(f, 1):
                matches = BUG_PATTERN.findall(line)
                for bug_num in matches:
                    results.append((f"BUG-{bug_num}", f"  L{line_no}: {line.strip()}"))
    except Exception as e:
        print(f"  [WARN] Could not read {filepath}: {e}")
    return results

def main():
    if not API_PATH.exists():
        print(f"Error: {API_PATH} does not exist")
        return

    print(f"Scanning {API_PATH} ...")
    all_findings = defaultdict(list)  # bug_num -> list of (file, line_info)

    go_files = list(API_PATH.rglob("*.go"))
    print(f"Found {len(go_files)} .go files\n")

    for i, filepath in enumerate(go_files, 1):
        findings = scan_file(filepath)
        if findings:
            rel_path = filepath.relative_to(API_PATH)
            for bug_num, line_info in findings:
                all_findings[bug_num].append((str(rel_path), line_info))
        
        if i % 100 == 0:
            print(f"  Processed {i}/{len(go_files)} files...")

    print(f"\n{'='*70}")
    print(f"BUG REFERENCES FOUND IN CODE: {len(all_findings)} unique bugs")
    print(f"{'='*70}\n")

    # Sort by bug number
    for bug_num in sorted(all_findings.keys(), key=lambda x: int(x.split('-')[1])):
        files = all_findings[bug_num]
        print(f"\n{bug_num} — found in {len(files)} location(s):")
        seen_files = set()
        for file_rel, line_info in files:
            if file_rel not in seen_files:
                print(f"  📄 {file_rel}")
                seen_files.add(file_rel)
            print(f"    {line_info}")

if __name__ == "__main__":
    main()
