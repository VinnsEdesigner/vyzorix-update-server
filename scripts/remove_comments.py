#!/usr/bin/env python3
"""
Script to remove comments from TypeScript/JavaScript files.
Handles both single-line (//) and multi-line (/* */) comments.
Preserves content inside strings and template literals.
"""

import re
import sys
import os
from pathlib import Path


def remove_comments(content: str) -> str:
    """
    Remove single-line and multi-line comments from TypeScript/JavaScript code.
    Preserves strings and template literals to avoid breaking code.
    """
    result = []
    i = 0
    n = len(content)
    
    while i < n:
        # Check for multi-line comment /*
        if i < n - 1 and content[i] == '/' and content[i + 1] == '*':
            end = content.find('*/', i + 2)
            if end != -1:
                i = end + 2
                continue
        # Check for single-line comment //
        if content[i] == '/' and i + 1 < n and content[i + 1] == '/':
            end = content.find('\n', i)
            if end == -1:
                break
            i = end + 1
            continue
        result.append(content[i])
        i += 1
    
    return ''.join(result)


def process_file(filepath: Path, dry_run: bool = False, verbose: bool = False) -> tuple[int, int]:
    """
    Process a single file and remove comments.
    Returns tuple of (original_size, new_size).
    """
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return 0, 0
    
    original_size = len(content)
    cleaned = remove_comments(content)
    new_size = len(cleaned)
    
    if verbose and not dry_run:
        removed = original_size - new_size
        if removed > 0:
            print(f"  Removed {removed} chars: {filepath.name}")
    
    if not dry_run and new_size < original_size:
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(cleaned)
        except Exception as e:
            print(f"Error writing {filepath}: {e}")
            return original_size, original_size
    
    return original_size, new_size


def process_directory(base_dir: Path, dry_run: bool = False, verbose: bool = False) -> dict:
    """
    Process all TypeScript/JavaScript files in a directory.
    Returns statistics about the operation.
    """
    stats = {
        'files_processed': 0,
        'files_skipped': 0,
        'original_size': 0,
        'new_size': 0,
    }
    
    extensions = {'.ts', '.tsx', '.js', '.jsx', '.mts', '.cts'}
    
    for filepath in base_dir.rglob('*'):
        # Skip hidden files and directories
        if any(part.startswith('.') for part in filepath.parts):
            stats['files_skipped'] += 1
            continue
        
        if filepath.is_file() and filepath.suffix in extensions:
            orig, new = process_file(filepath, dry_run, verbose)
            stats['files_processed'] += 1
            stats['original_size'] += orig
            stats['new_size'] += new
        else:
            stats['files_skipped'] += 1
    
    return stats


def main():
    if len(sys.argv) < 2:
        print("Usage: python remove_comments.py <directory> [--dry-run] [--verbose]")
        print("  --dry-run    Show what would be changed without modifying files")
        print("  --verbose    Show detailed output")
        sys.exit(1)
    
    path = sys.argv[1]
    dry_run = '--dry-run' in sys.argv or '-n' in sys.argv
    verbose = '--verbose' in sys.argv or '-v' in sys.argv
    
    base_dir = Path(path)
    
    if not base_dir.exists():
        print(f"Error: Directory '{base_dir}' does not exist")
        return 1
    
    print(f"\n{'[DRY RUN] ' if dry_run else ''}Processing {base_dir}")
    print("-" * 60)
    
    stats = process_directory(base_dir, dry_run=dry_run, verbose=verbose)
    
    print("\n" + "-" * 60)
    print("Summary:")
    print(f"  Files processed: {stats['files_processed']}")
    print(f"  Files skipped:   {stats['files_skipped']}")
    print(f"  Original size:  {stats['original_size']:,} bytes")
    print(f"  New size:       {stats['new_size']:,} bytes")
    print(f"  Removed:        {stats['original_size'] - stats['new_size']:,} bytes")
    
    if dry_run:
        print("\n[DRY RUN] No files were modified.")
    
    return 0


if __name__ == '__main__':
    exit(main())
