#!/usr/bin/env python3
"""Live cross-tenant isolation probes against a running server.

Tenant A (key vxyz_tenantA_...) is a member of org-aaaa-1111 only.
Tenant B (key vxyz_tenantB_...) is a member of org-bbbb-2222 only.

We authenticate as A and try to read B's data. Every probe must be DENIED.
"""
import base64
import hashlib
import hmac
import json
import sys
import time
import urllib.error
import urllib.request
import uuid

BASE = "http://localhost:3000"
KEY_A = "vxyz_tenantA_1234567890"
KEY_B = "vxyz_tenantB_1234567890"
ORG_A = "org-aaaa-1111"
ORG_B = "org-bbbb-2222"

# signing_secret = sha512(fullKey).hexdigest  (see deriveAPIKeySigningSecret)
SECRET_A = hashlib.sha512(KEY_A.encode()).hexdigest()
SECRET_B = hashlib.sha512(KEY_B.encode()).hexdigest()

results = []


def sign(method, path, body: bytes, secret_hex: str):
    nonce = uuid.uuid4().hex
    ts = str(int(time.time() * 1000))
    msg = f"{method}\n{path}\n{nonce}\n{ts}\n".encode() + body
    sig = base64.b64encode(
        hmac.new(secret_hex.encode(), msg, hashlib.sha512).digest()
    ).decode()
    return {
        "X-Vyzorix-Nonce": nonce,
        "X-Vyzorix-Timestamp": ts,
        "X-Vyzorix-Signature": sig,
    }


def request(method, path, key, secret_hex, body=None, extra_headers=None):
    raw = b"" if body is None else json.dumps(body).encode()
    headers = {"X-API-Key": key, "Content-Type": "application/json"}
    headers.update(sign(method, path, raw, secret_hex))
    if extra_headers:
        headers.update(extra_headers)
    req = urllib.request.Request(BASE + path, data=raw or None, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            return r.status, r.read().decode()[:400]
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode()[:400]
    except Exception as e:  # noqa: BLE001
        return -1, str(e)


def record(name, expect_denied, status, body):
    # GraphQL returns HTTP 200 even for authorization errors; the denial lives in
    # the JSON "errors" array with data nulled. REST denials use a 4xx status.
    denied = False
    detail = ""
    try:
        parsed = json.loads(body)
        if isinstance(parsed, dict):
            if parsed.get("errors"):
                # GraphQL: denial lives in the errors array, data nulled.
                denied = True
                detail = parsed["errors"][0].get("message", "")
                if any(v is not None for v in (parsed.get("data") or {}).values()):
                    denied = False
                    detail = "ERROR: response carried data alongside the error"
            elif parsed.get("error"):
                # REST: single error envelope.
                denied = True
                detail = parsed["error"].get("message", "")
    except (ValueError, TypeError):
        pass

    if not denied and 400 <= status < 500:
        denied = True
        detail = detail or f"http {status}"

    ok = (denied == expect_denied)
    results.append((name, expect_denied, status, ok, detail))
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: denied={denied} (expect_denied={expect_denied}) http={status}")
    if detail:
        print(f"        {detail[:160]}")
    if not ok:
        print(f"        body={body[:200]}")


def gql(org_in_path, key, secret_hex, query, variables=None):
    body = {"query": query}
    if variables:
        body["variables"] = variables
    return request("POST", f"/{org_in_path}/graphql", key, secret_hex, body)


print("=== sanity: tenant A CAN read its own org (must be allowed) ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ alertRules(organizationId: "%s") { id } }' % ORG_A)
print(f"    same-org alertRules -> {s} {b[:120]}")
if not (200 <= s < 300):
    print("    !! baseline failed; aborting (harness/auth issue, not a security result)")
    sys.exit(3)

print()
print("=== PROBE 1: A reads B's alert rules by passing B's organizationId ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ alertRules(organizationId: "%s") { id name } }' % ORG_B)
record("gql alertRules cross-org", True, s, b)

print()
print("=== PROBE 2: A reads B's service accounts ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ serviceAccounts(organizationId: "%s") { id name } }' % ORG_B)
record("gql serviceAccounts cross-org", True, s, b)

print()
print("=== PROBE 3: A reads B's contact points ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ contactPoints(organizationId: "%s") { id name } }' % ORG_B)
record("gql contactPoints cross-org", True, s, b)

print()
print("=== PROBE 4: A reads B's annotations ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ annotations(organizationId: "%s") { id title } }' % ORG_B)
record("gql annotations cross-org", True, s, b)

print()
print("=== PROBE 5: A mutates B's org (createAlertRule with B's organizationId) ===")
s, b = gql(ORG_A, KEY_A, SECRET_A,
           'mutation { createAlertRule(organizationId: "%s", name: "pwn", metric: "riskScore", '
           'condition: "gt", threshold: 1, enabled: true) { id name } }' % ORG_B)
record("gql createAlertRule cross-org", True, s, b)

print()
print("=== PROBE 6: A reads REST alerts for B via X-Organization-ID ===")
s, b = request("GET", "/v1/alerts/rules", KEY_A, SECRET_A, extra_headers={"X-Organization-ID": ORG_B})
record("rest alerts cross-org header", True, s, b)

print()
print("=== PROBE 7 (control): B reads B's own org - must be ALLOWED ===")
s, b = gql(ORG_B, KEY_B, SECRET_B, '{ alertRules(organizationId: "%s") { id } }' % ORG_B)
record("gql B reads own org (control)", False, s, b)

print()
print("=== PROBE 8 (control): A reads A's own org - must be ALLOWED ===")
s, b = gql(ORG_A, KEY_A, SECRET_A, '{ alertRules(organizationId: "%s") { id } }' % ORG_A)
record("gql A reads own org (control)", False, s, b)

failed = [r for r in results if not r[3]]
print()
print(f"=== {len(results) - len(failed)}/{len(results)} probes behaved correctly ===")
for name, _, status, _, _ in failed:
    print(f"    FAILED: {name} (status={status})")
sys.exit(1 if failed else 0)
