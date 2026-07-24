#!/bin/bash
# Architecture Analysis Script - Comprehensive
# 
# ARCHITECTURE BOUNDARY:
# 
#   NEW (CLEAN) structure - MUST use only NEW imports                 
#     • internal/ws/                                                 
#     • internal/infrastructure/ (except internal/api/ - OLD)         
#     • internal/domain/                                             
#     • internal/application/                                         
#     • cmd/api/                                                     
# 
# 
#   OLD (FLAT) structure - Being replaced/migrated                    
#     • pkg/ (all subpackages)                                       
#     • internal/api/handlers/*.go (root level flat handlers)         
#     • internal/audit/logger.go (OLD version)                       
#     • internal/command_signer.go (OLD version)                      
#     • internal/email.go                                            
# 
#
# LEGITIMATE:
#   • NEW files CAN import pkg/config (API layer concern)
#   • OLD files importing pkg/ is expected - needs migration

set -e
cd "$(dirname "$0")/.."

echo ""
echo "                    COMPREHENSIVE ARCHITECTURE ANALYSIS                              "
echo ""
echo ""

# Define NEW directories (clean architecture)
NEW_DIRS="internal/ws internal/infrastructure internal/domain internal/application cmd/api"

# Define OLD directories/packages
OLD_PATTERNS="vyzorix/apps/api/pkg vyzorix/apps/api/internal/api/handlers"

# Legitimate imports (pkg/config is OK for API layer)
LEGITIMATE="vyzorix/apps/api/pkg/config"

# ============================================================
# STEP 1: Check TRUE NEW files importing from ANY OLD structure
# ============================================================
echo ""
echo " STEP 1: NEW structure files importing from OLD structure (PROBLEM!)"
echo ""
echo ""
echo "NEW structure: internal/ws/, internal/infrastructure/, internal/domain/, cmd/api/"
echo "OLD structure: pkg/*, internal/api/handlers/*"
echo ""

VIOLATION_FOUND=0
LEGITIMATE_FOUND=0
for dir in $NEW_DIRS; do
    if [ -d "$dir" ]; then
        while IFS= read -r f; do
            # Skip test files
            if echo "$f" | grep -q "_test.go"; then continue; fi
            
            # Check for violations (OLD imports)
            has_old_import=0
            has_violation=0
            violations=""
            legitimates=""
            
            for old_pat in $OLD_PATTERNS; do
                if grep -q "$old_pat" "$f" 2>/dev/null; then
                    has_old_import=1
                    # Check what exactly is imported
                    imports=$(grep -h "$old_pat" "$f" 2>/dev/null)
                    
                    # Check if all imports are legitimate (pkg/config)
                    if echo "$imports" | grep -q "pkg/config"; then
                        if echo "$imports" | grep -qv "pkg/models\|pkg/storage\|pkg/crypto\|pkg/logging"; then
                            # Only pkg/config imports - legitimate
                            legitimates="$imports"
                        else
                            has_violation=1
                            violations="$imports"
                        fi
                    else
                        has_violation=1
                        violations="$imports"
                    fi
                fi
            done
            
            if [ $has_violation -eq 1 ]; then
                VIOLATION_FOUND=1
                echo " VIOLATION: $f"
                echo "$violations" | sed 's/^/   → /'
            fi
        done < <(find "$dir" -name "*.go" 2>/dev/null)
    fi
done

if [ $VIOLATION_FOUND -eq 0 ]; then
    echo " All NEW structure files are CLEAN!"
    echo "   No violations found in: $NEW_DIRS"
fi

echo ""

# ============================================================
# STEP 2: Check for OLD files that should be migrated
# ============================================================
echo ""
echo " STEP 2: OLD structure files importing from OLD pkg/ (EXPECTED - needs migration)"
echo ""
echo ""
echo "These files are in the OLD flat structure and are expected to import from pkg/"
echo ""

# Check OLD files
OLD_FILES=""
# Flat handlers
for f in internal/api/handlers/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test.go"; then
        OLD_FILES="$OLD_FILES $f"
    fi
done

# Other OLD root-level files
for f in internal/command_signer.go internal/email.go; do
    if [ -f "$f" ]; then
        OLD_FILES="$OLD_FILES $f"
    fi
done

for f in $OLD_FILES; do
    HANDLER_NAME=$(basename "$f")
    echo " $HANDLER_NAME (needs migration to NEW structure)"
    grep 'vyzorix/apps/api/pkg' "$f" 2>/dev/null | sed 's/^/   → /'
    echo ""
done

echo ""

# ============================================================
# STEP 3: Count all OLD → OLD imports (expected)
# ============================================================
echo ""
echo " STEP 3: Summary by Package"
echo ""
echo ""

# Count imports in NEW structure from OLD packages
echo ""
echo " Import Category                                  Count                                  "
echo ""

# NEW → pkg/ (violations - exclude legitimate pkg/config)
NEW_PKG_VIO=$(find $NEW_DIRS -name "*.go" 2>/dev/null | grep -v "_test.go" | while read f; do 
    if grep -q "vyzorix/apps/api/pkg/models\|vyzorix/apps/api/pkg/storage\|vyzorix/apps/api/pkg/crypto\|vyzorix/apps/api/pkg/logging" "$f" 2>/dev/null; then
        echo "$f"
    fi
done | wc -l)
printf " NEW structure → pkg/ (VIOLATIONS)              %38d \n" "$NEW_PKG_VIO"

# NEW → internal/api/handlers (violations)
NEW_HANDLERS_VIO=$(find $NEW_DIRS -name "*.go" 2>/dev/null | grep -v "_test.go" | while read f; do 
    if grep -q "vyzorix/apps/api/internal/api/handlers" "$f" 2>/dev/null; then
        echo "$f"
    fi
done | wc -l)
printf " NEW structure → internal/api/handlers (VIO)  %38d \n" "$NEW_HANDLERS_VIO"

# OLD → pkg/ (expected, needs migration)
OLD_PKG=$(echo "$OLD_FILES" | tr ' ' '\n' | while read f; do if [ -f "$f" ]; then grep -l "vyzorix/apps/api/pkg" "$f" 2>/dev/null; fi; done | wc -l)
printf " OLD structure → pkg/ (needs migration)        %38d \n" "$OLD_PKG"

echo ""
echo ""

# ============================================================
# STEP 4: Final Status
# ============================================================
echo ""
TOTAL_VIO=$((NEW_PKG_VIO + NEW_HANDLERS_VIO))
if [ $VIOLATION_FOUND -eq 0 ]; then
    echo " NEW structure is CLEAN - no violations!"
    echo ""
    echo "All files in internal/ws/, internal/infrastructure/, internal/domain/,"
    echo "internal/application/, cmd/api/ use only NEW architecture imports."
else
    echo "  NEW structure has $VIOLATION_FOUND violations that need fixing!"
fi
echo ""
echo "OLD files (internal/api/handlers/*.go, internal/command_signer.go, internal/email.go)"
echo "are expected to import from pkg/ and need gradual migration."
echo ""
