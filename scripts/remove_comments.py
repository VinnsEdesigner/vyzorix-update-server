#!/usr/bin/env python3
import re
import sys
import os

def remove_comments(content):
    content = re.sub(r'/\*[\s\S]*?\*/', '', content)
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    return content

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    original = content
    content = remove_comments(content)
    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Cleaned: {filepath}")
        return True
    return False

def process_directory(dirpath, extensions=['.ts', '.tsx', '.js', '.jsx']):
    count = 0
    for root, dirs, files in os.walk(dirpath):
        if 'node_modules' in root:
            continue
        for file in files:
            if any(file.endswith(ext) for ext in extensions):
                filepath = os.path.join(root, file)
                if process_file(filepath):
                    count += 1
    return count

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python remove_comments.py <file_or_directory>")
        sys.exit(1)
    path = sys.argv[1]
    if os.path.isfile(path):
        process_file(path)
    elif os.path.isdir(path):
        count = process_directory(path)
        print(f"Cleaned {count} files")
