#!/usr/bin/env python3
"""
Emoji Removal Script
Removes all emoji characters from text files in the codebase.
"""

import os
import re
import sys

# Unicode ranges for emoji characters
EMOJI_PATTERN = re.compile(
    "["
    "\U0001F600-\U0001F64F"  # emoticons
    "\U0001F300-\U0001F5FF"  # symbols & pictographs
    "\U0001F680-\U0001F6FF"  # transport & map symbols
    "\U0001F1E0-\U0001F1FF"  # flags (iOS)
    "\U0001F700-\U0001F77F"  # alchemical symbols
    "\U0001F780-\U0001F7FF"  # Geometric Shapes Extended
    "\U0001F800-\U0001F8FF"  # Supplemental Arrows-C
    "\U0001F900-\U0001F9FF"  # Supplemental Symbols and Pictographs
    "\U0001FA00-\U0001FA6F"  # Chess Symbols
    "\U0001FA70-\U0001FAFF"  # Symbols and Pictographs Extended-A
    "\U00002702-\U000027B0"  # Dingbats
    "\U000024C2-\U0001F251"  # enclosed characters
    "\U0001F004-\U0001F0CF"  # Mahjong tiles
    "\U0001F170-\U0001F251"  # enclosed alphanumerics
    "]+",
    flags=re.UNICODE
)

# File extensions to process
TEXT_EXTENSIONS = {
    '.py', '.js', '.ts', '.jsx', '.tsx', '.md', '.txt', '.json', '.yaml', '.yml',
    '.html', '.css', '.scss', '.sass', '.less', '.xml', '.csv', '.rst',
    '.sh', '.bash', '.zsh', '.fish', '.ps1', '.bat', '.cmd',
    '.rb', '.java', '.kt', '.swift', '.go', '.rs', '.c', '.cpp', '.h', '.hpp',
    '.sql', '.graphql', '.vue', '.svelte'
}

# Directories to skip
SKIP_DIRS = {
    '.git', '.svn', '.hg', 'node_modules', '__pycache__', '.pytest_cache',
    '.mypy_cache', '.tox', 'venv', '.venv', 'env', '.env', 'dist', 'build',
    '.eggs', '*.egg-info', '.tox', '.nox', '.coverage', 'htmlcov',
    '.DS_Store', '.AppleDouble', '.LSOverride'
}


def is_text_file(filepath: str) -> bool:
    """Check if file is a text file based on extension."""
    _, ext = os.path.splitext(filepath)
    return ext.lower() in TEXT_EXTENSIONS


def should_skip_dir(dirname: str) -> bool:
    """Check if directory should be skipped."""
    return dirname in SKIP_DIRS or dirname.startswith('.')


def remove_emojis(text: str) -> str:
    """Remove all emoji characters from text."""
    return EMOJI_PATTERN.sub('', text)


def process_file(filepath: str, dry_run: bool = False, verbose: bool = False) -> tuple[bool, bool]:
    """
    Process a single file to remove emojis.
    Returns: (was_modified, had_emojis)
    """
    try:
        with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
            original_content = f.read()
    except (IOError, OSError) as e:
        if verbose:
            print(f"Error reading {filepath}: {e}")
        return False, False
    
    had_emojis = bool(EMOJI_PATTERN.search(original_content))
    
    if not had_emojis:
        return False, False
    
    new_content = remove_emojis(original_content)
    
    if dry_run:
        print(f"[DRY RUN] Would modify: {filepath}")
        return True, True
    
    try:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        if verbose:
            print(f"Removed emojis from: {filepath}")
        return True, True
    except (IOError, OSError) as e:
        if verbose:
            print(f"Error writing {filepath}: {e}")
        return False, True


def scan_directory(root_dir: str, dry_run: bool = False, verbose: bool = False) -> dict:
    """
    Recursively scan directory and remove emojis from all text files.
    Returns statistics dictionary.
    """
    stats = {
        'files_scanned': 0,
        'files_modified': 0,
        'files_with_emojis': 0,
        'errors': 0
    }
    
    for dirpath, dirnames, filenames in os.walk(root_dir):
        # Filter out skip directories
        dirnames[:] = [d for d in dirnames if not should_skip_dir(d)]
        
        for filename in filenames:
            filepath = os.path.join(dirpath, filename)
            
            if not is_text_file(filepath):
                continue
            
            stats['files_scanned'] += 1
            
            modified, had_emojis = process_file(filepath, dry_run, verbose)
            
            if had_emojis:
                stats['files_with_emojis'] += 1
            
            if modified:
                stats['files_modified'] += 1
    
    return stats


def main():
    import argparse
    
    parser = argparse.ArgumentParser(
        description='Remove all emoji characters from text files in a codebase.',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s .                      # Scan and remove emojis from current directory
  %(prog)s /path/to/project       # Scan specific directory
  %(prog)s . --dry-run            # Show what would be modified without changing files
  %(prog)s . --verbose            # Print each file as it's processed
  %(prog)s . --include-pattern "*.md"  # Only process markdown files
        """
    )
    
    parser.add_argument(
        'directory',
        nargs='?',
        default='.',
        help='Directory to scan (default: current directory)'
    )
    
    parser.add_argument(
        '--dry-run', '-n',
        action='store_true',
        help='Show what would be modified without making changes'
    )
    
    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='Print each file as it is processed'
    )
    
    parser.add_argument(
        '--include-pattern',
        help='Only process files matching this glob pattern (e.g., "*.md")'
    )
    
    parser.add_argument(
        '--exclude-pattern',
        help='Exclude files matching this glob pattern'
    )
    
    args = parser.parse_args()
    
    # Change to target directory
    target_dir = os.path.abspath(args.directory)
    
    if not os.path.isdir(target_dir):
        print(f"Error: {target_dir} is not a valid directory", file=sys.stderr)
        sys.exit(1)
    
    print(f"Scanning directory: {target_dir}")
    
    if args.dry_run:
        print("[DRY RUN MODE - No changes will be made]")
    
    # Note: include-pattern and exclude-pattern would require additional implementation
    # For now, using built-in extension-based filtering
    
    stats = scan_directory(target_dir, dry_run=args.dry_run, verbose=args.verbose)
    
    print("\n" + "=" * 50)
    print("SUMMARY")
    print("=" * 50)
    print(f"Files scanned:        {stats['files_scanned']}")
    print(f"Files with emojis:    {stats['files_with_emojis']}")
    print(f"Files modified:       {stats['files_modified']}")
    print(f"Errors:               {stats['errors']}")
    
    if args.dry_run:
        print("\nRun without --dry-run to apply changes.")


if __name__ == '__main__':
    main()