#!/bin/bash
# Scan remaining internal/auth/ files

cd "$(dirname "$0")/.."

echo ""
echo "                    SCANNING: REMAINING internal/auth/ FILES                                          "
echo ""
echo ""

echo ""
echo "1. REMAINING FILES IN internal/auth/"
echo ""
echo ""

echo "Files:"
for f in internal/auth/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
        types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
        echo "  $(basename $f): $types types, $funcs funcs"
    fi
done

echo ""
echo ""
echo "2. MIGRATION STATUS"
echo ""
echo ""

echo ""
echo " File                                               Migration Status                        "
echo ""

# Check each file
for f in internal/auth/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        name=$(basename "$f")
        
        case "$name" in
            "jwt.go")
                if grep -rq "JWTManager\|Generate\|Verify" internal/application/auth/ 2>/dev/null; then
                    echo " $name                              In application/auth/           "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "google_token.go"|"github.go")
                if grep -rq "Google\|GitHub\|OAuth" internal/domain/oauth/ 2>/dev/null || grep -rq "Google\|GitHub" internal/application/auth/ 2>/dev/null; then
                    echo " $name                              In domain/application/        "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "session.go")
                if grep -rq "SessionManager" internal/application/ 2>/dev/null || grep -rq "SessionManager" internal/infrastructure/storage/session.go 2>/dev/null; then
                    echo " $name                              In application/storage       "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "validate.go")
                if grep -rq "Validate" internal/domain/ 2>/dev/null; then
                    echo " $name                              In domain layer              "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "password.go")
                if grep -rq "HashPassword\|VerifyPassword" internal/infrastructure/auth/ 2>/dev/null; then
                    echo " $name                              In infrastructure/auth/      "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "lockout.go"|"totp.go"|"totp_qr.go"|"origin.go"|"ratelimit.go")
                if grep -rq "Lockout\|TOTP\|Origin\|RateLimit" internal/infrastructure/security/ 2>/dev/null; then
                    echo " $name                              In infrastructure/security/  "
                else
                    echo " $name                               Needs migration              "
                fi
                ;;
            "revocation.go"|"request_signer.go")
                echo " $name                               Check middleware/security   "
                ;;
            *)
                echo " $name                               Unknown                       "
                ;;
        esac
    fi
done

echo ""

echo ""
echo ""
echo "3. KEY LOGIC SUMMARY"
echo ""
echo ""

echo "JWT/Token Management:"
grep "^func " internal/auth/jwt.go 2>/dev/null | head -5

echo ""
echo "OAuth:"
grep "^func " internal/auth/google_token.go 2>/dev/null | grep -i "google\|oauth" | head -5
grep "^func " internal/auth/github.go 2>/dev/null | grep -i "github\|oauth" | head -5

echo ""
echo "Session:"
grep "^func " internal/auth/session.go 2>/dev/null | head -5

echo ""
echo "Validation:"
grep "^func " internal/auth/validate.go 2>/dev/null | head -10

echo ""
echo ""
echo "4. FILES TO MIGRATE"
echo ""
echo ""

echo "Files needing migration to infrastructure/security/ or application/:"
for f in internal/auth/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        name=$(basename "$f")
        migrated="no"
        
        # Check if already migrated
        case "$name" in
            "password.go") grep -q "HashPassword" internal/infrastructure/auth/argon2_hasher.go 2>/dev/null && migrated="yes" ;;
            "lockout.go"|"totp.go"|"totp_qr.go"|"origin.go"|"ratelimit.go")
                grep -rq "$(echo $name | sed 's/.go//')" internal/infrastructure/security/ 2>/dev/null && migrated="yes" ;;.
            "jwt.go") grep -rq "JWTManager" internal/application/auth/ 2>/dev/null && migrated="yes" ;;
            "google_token.go"|"github.go") grep -rq "Google\|GitHub" internal/domain/oauth/ 2>/dev/null && migrated="yes" ;;
        esac
        
        if [ "$migrated" = "no" ]; then
            echo "    $name"
        fi
    fi
done

echo ""
echo ""
