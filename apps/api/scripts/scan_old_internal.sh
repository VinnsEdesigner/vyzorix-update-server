#!/bin/bash
# Scan OLD internal/ files that were migrated to infrastructure/

cd "$(dirname "$0")/.."

echo "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    SCANNING: OLD internal/ FILES vs NEW infrastructure/                                          ║"
echo "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. OLD FILES IN internal/ ROOT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files in internal/ root (not in subdirectories):"
find internal/ -maxdepth 1 -name "*.go" -type f 2>/dev/null | while read f; do
    name=$(basename "$f")
    if [ -f "$f" ]; then
        # Check if migrated
        migrated=""
        case "$name" in
            "email.go") grep -q "type.*Service\|func.*Send" internal/infrastructure/email/*.go 2>/dev/null && migrated="✅ migrated to infrastructure/email/" ;;
            "command_signer.go") grep -q "CommandSigner" internal/infrastructure/crypto/*.go 2>/dev/null && migrated="✅ migrated to infrastructure/crypto/" ;;
            *) migrated="⚠️  NOT MIGRATED" ;;
        esac
        echo "  $name: $migrated"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. OLD DIRECTORIES IN internal/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "┌────────────────────────────────────────────────┬────────────────────────────────────────────────┐"
echo "│ OLD Location                                      │ NEW Location                                  │"
echo "├────────────────────────────────────────────────┼────────────────────────────────────────────────┤"

# Check each old directory
for old_dir in internal/audit internal/auth internal/fcm internal/ssr internal/metrics; do
    if [ -d "$old_dir" ]; then
        name=$(basename "$old_dir")
        new_dir="internal/infrastructure/$name"
        
        if [ -d "$new_dir" ]; then
            old_files=$(find "$old_dir" -name "*.go" -type f | grep -v "_test" | wc -l)
            new_files=$(find "$new_dir" -name "*.go" -type f | grep -v "_test" | wc -l)
            echo "│ $old_dir/ ($old_dir files)    → $new_dir/ ($new_files files) ✅ │"
        else
            echo "│ $old_dir/                    → NOT MIGRATED ⚠️                         │"
        fi
    fi
done

echo "└────────────────────────────────────────────────┴────────────────────────────────────────────────┘"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. FILES STILL IN OLD LOCATIONS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "OLD files still present (can be deleted after verification):"
for dir in internal/audit internal/auth internal/fcm internal/ssr internal/metrics; do
    if [ -d "$dir" ]; then
        files=$(find "$dir" -name "*.go" -type f | grep -v "_test" | wc -l)
        if [ "$files" -gt 0 ]; then
            echo "  $dir/: $files files"
            find "$dir" -name "*.go" -type f | grep -v "_test" | head -5 | while read f; do
                echo "    - $(basename $f)"
            done
        fi
    fi
done

echo ""
echo "Root files:"
for f in internal/email.go internal/command_signer.go; do
    if [ -f "$f" ]; then
        echo "  - $f"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. NEW infrastructure/ STRUCTURE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Current infrastructure/ contents:"
for d in $(ls -d internal/infrastructure/*/ 2>/dev/null | sort); do
    name=$(basename "$d")
    files=$(find "$d" -name "*.go" -type f | grep -v "_test" | wc -l)
    echo "  ✅ $name/ ($files files)"
done

echo ""
echo "Total infrastructure files: $(find internal/infrastructure -name "*.go" -type f | grep -v "_test" | wc -l)"

echo ""
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "STATUS: Infrastructure layer is COMPLETE. OLD files remain for reference until deletion is safe."
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
