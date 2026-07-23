#!/usr/bin/env python3
"""Add trailing dots to // comments that don't have them.
Skips: empty comments, comments at file start (line 1), URL-like // patterns.
"""
import os
import re
import sys

IGNORE_DIRS = {'.git', '.svn', 'node_modules', 'vendor', '__pycache__',
               'dist', 'build', 'target', '.venv', 'venv'}
IGNORE_EXT = {'.zip', '.tar', '.gz', '.bz2', '.xz',
              '.png', '.jpg', '.jpeg', '.gif', '.ico', '.woff', '.woff2',
              '.pdf', '.doc', '.docx'}
IGNORE_FILES = {'go.sum', 'go.mod', 'Makefile'}


def ends_with_dot(s):
    return s.rstrip().endswith('.')


def is_url_line(line):
    """Skip lines with URL-like // patterns."""
    if '://' in line:
        return True
    if '"//' in line or "'//" in line:
        return True
    if '$$(' in line:  # Makefile variables
        return True
    return False


def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
        lines = f.readlines()

    modified = False
    new_lines = []

    for i, line in enumerate(lines, 1):
        if '//' in line and not is_url_line(line):
            # Find the comment part
            idx = line.find('//')
            comment = line[idx:].rstrip()
            
            # Skip empty comments
            stripped = comment.lstrip()
            if stripped in ('//', '/*', ''):
                new_lines.append(line)
                continue
            
            # Skip line 1 (package/file comments)
            if i == 1:
                new_lines.append(line)
                continue
            
            # Add dot if missing
            if not ends_with_dot(comment):
                # Find last non-whitespace before //
                prefix = line[:idx]
                # Insert dot before trailing whitespace
                new_line = prefix + comment + '.\n'
                new_lines.append(new_line)
                modified = True
            else:
                new_lines.append(line)
        else:
            new_lines.append(line)

    if modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)
        return True
    return False


def should_ignore(path):
    name = os.path.basename(path)
    if not name:
        return False
    if name in {'add_dots.py', 'add_dots', 'strip_bugs', 'strip_bugs.cpp'}:
        return True
    if name in IGNORE_FILES:
        return True
    if name in IGNORE_DIRS:
        return True
    if os.path.splitext(name)[1] in IGNORE_EXT:
        return True
    if name.startswith('.'):
        return True
    return False


def walk_dir(root):
    changed = []
    for dirpath, dirnames, filenames in os.walk(root):
        # Remove ignored dirs from traversal
        dirnames[:] = [d for d in dirnames if not should_ignore(os.path.join(dirpath, d))]

        for fname in filenames:
            fpath = os.path.join(dirpath, fname)
            if should_ignore(fpath):
                continue
            if process_file(fpath):
                print(f"[DOTTED] {fpath}")
                changed.append(fpath)
    return changed


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <root_dir>")
        return 1

    root = sys.argv[1]
    if not os.path.isdir(root):
        print(f"Error: not a valid directory: {root}")
        return 1

    changed = walk_dir(root)
    print(f"\nDone. {len(changed)} file(s) modified.")
    return 0


if __name__ == '__main__':
    sys.exit(main())
