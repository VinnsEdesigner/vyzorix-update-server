#!/usr/bin/env python3
"""Live cross-tenant probes for the GraphQL subscription (WebSocket) path.

Tenant A (key vxyz_tenantA_...) is a member of org-aaaa-1111 only.

The connection authenticates against the org in the URL. The probes then try to
subscribe with a *different* org / device supplied in the subscription variables.
Every such attempt must be rejected.

Requires: pip install websockets
"""
import asyncio
import json
import sys

import websockets

BASE_WS = "ws://localhost:3000"
KEY_A = "vxyz_tenantA_1234567890"
ORG_A = "org-aaaa-1111"
ORG_B = "org-bbbb-2222"

results = []


async def open_conn(org):
    return await websockets.connect(
        f"{BASE_WS}/{org}/graphql/ws",
        additional_headers={"X-API-Key": KEY_A},
        open_timeout=15,
    )


async def subscribe(ws, sub_id, query, variables):
    await ws.send(json.dumps({
        "type": "subscribe",
        "id": sub_id,
        "payload": {"query": query, "variables": variables},
    }))
    # First message back is the ack/next or an error for this id.
    raw = await asyncio.wait_for(ws.recv(), timeout=10)
    return json.loads(raw)


def record(name, expect_denied, msg):
    denied = msg.get("type") == "error"
    detail = ""
    if denied:
        payload = msg.get("payload") or {}
        detail = payload.get("message", "") if isinstance(payload, dict) else str(payload)
    ok = (denied == expect_denied)
    results.append((name, ok, detail or msg.get("type")))
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: denied={denied} (expect_denied={expect_denied})")
    if detail:
        print(f"        {detail[:160]}")
    if not ok:
        print(f"        msg={json.dumps(msg)[:220]}")


async def main():
    print("=== control: A connects to its own org and subscribes to its own org ===")
    ws = await open_conn(ORG_A)
    await ws.send(json.dumps({"type": "connection_init"}))
    ack = json.loads(await asyncio.wait_for(ws.recv(), timeout=10))
    print(f"    handshake -> {ack.get('type')}")
    if ack.get("type") != "connection_ack":
        print("    !! handshake failed; aborting (harness issue)")
        return 3

    msg = await subscribe(
        ws, "own", "subscription { organizationEvent(orgId: \"%s\") { __typename } }" % ORG_A,
        {"orgId": ORG_A},
    )
    record("ws own-org organizationEvent (control)", False, msg)
    await ws.close()

    print()
    print("=== PROBE 1: A (org A) subscribes to org B's organizationEvent ===")
    ws = await open_conn(ORG_A)
    await ws.send(json.dumps({"type": "connection_init"}))
    await asyncio.wait_for(ws.recv(), timeout=10)
    msg = await subscribe(
        ws, "x1", "subscription { organizationEvent(orgId: \"%s\") { __typename } }" % ORG_B,
        {"orgId": ORG_B},
    )
    record("ws organizationEvent cross-org", True, msg)
    await ws.close()

    print()
    print("=== PROBE 2: A subscribes to org B's memberEvent ===")
    ws = await open_conn(ORG_A)
    await ws.send(json.dumps({"type": "connection_init"}))
    await asyncio.wait_for(ws.recv(), timeout=10)
    msg = await subscribe(
        ws, "x2", "subscription { memberEvent(orgId: \"%s\") { __typename } }" % ORG_B,
        {"orgId": ORG_B},
    )
    record("ws memberEvent cross-org", True, msg)
    await ws.close()

    print()
    print("=== PROBE 3: A connects to org B's URL (not a member) ===")
    try:
        ws = await open_conn(ORG_B)
        await ws.send(json.dumps({"type": "connection_init"}))
        raw = await asyncio.wait_for(ws.recv(), timeout=10)
        # If we get here the handshake was accepted - that is the failure.
        record("ws connect to non-member org", True, {"type": "accepted", "payload": {}})
        print(f"        unexpectedly connected: {raw[:120]}")
        await ws.close()
    except Exception as e:  # noqa: BLE001
        record("ws connect to non-member org", True, {"type": "error", "payload": {"message": str(e)}})

    failed = [r for r in results if not r[1]]
    print()
    print(f"=== {len(results) - len(failed)}/{len(results)} probes behaved correctly ===")
    for name, _, detail in failed:
        print(f"    FAILED: {name} ({detail[:80]})")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
