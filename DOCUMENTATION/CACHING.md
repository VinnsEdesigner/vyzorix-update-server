# Caching

Some answers are expensive to compute and don't change between two requests a
second apart. The dashboard's aggregate stats, a telemetry-history query, the
result of a cross-resource search — recomputing them on every keystroke is waste.
The cache holds those answers for a short, configurable time so the second ask is
free.

It is a plain in-memory TTL cache. No external store, no eviction daemon, no
serialization. Entries expire lazily — a key is checked on read and dropped if
it's stale — and every section reports its own hit and miss counts.


---

## The shape of it

A cache with named **sections**. A section is a scope that shares the cache's
expiry but keeps its own hit/miss counters, so you can see which parts of the
system are actually benefiting.

```
Cache (default TTL)
├── section "dashboard"   → hit/miss counters
├── section "telemetry"   → hit/miss counters
└── section "search"      → hit/miss counters
```

Every entry is a key → value pair with an expiry. A read either returns the value
(hit) or reports a miss. A section can also set an entry with its own TTL when
the default doesn't fit.

---

## What it's for

| Section | Caches | Why |
|:--------|:-------|:----|
| `dashboard` | Aggregate dashboard stats. | The dashboard polls; the aggregate is the expensive part. |
| `telemetry` | Telemetry-history query results. | History queries scan a lot of rows for an answer that rarely changes mid-session. |
| `search` | Cross-resource search results. | Search runs several queries per keystroke; the same query shouldn't pay twice. |

### Invalidation

A cache is only useful if it's correct. When something changes the underlying
data, the relevant section is dropped:

- When an **alert fires or resolves**, the dashboard section is invalidated, so
  the dashboard doesn't show stale counts next to a fresh alert.
- When a **rollout starts**, the dashboard section is invalidated for the same
  reason.

This keeps the cached dashboard honest: it may be a few seconds stale, but it
never contradicts an alert the operator can see.

---

## Hit and miss metrics

Each section emits a `cache_operations_total` counter labeled by section and
outcome (`hit` / `miss`). That's how you answer "is this cache earning its
keep?" — a section with a high miss rate is caching something that changes too
often to bother.

---

## Configuration

| Variable | Default | Meaning |
|:---------|:--------|:--------|
| `CACHE_TTL_SECONDS` | `30` | Default lifetime of a cached entry. Set to `0` to disable caching entirely. |

Per-entry TTLs can override the default with `SetTTL` when a specific answer
needs a different lifetime.

---

## Why in-memory, not Redis

The cache exists to save CPU and query time on a single instance, not to share
state across a fleet. There's nothing in it that can't be recomputed — if the
process restarts, the cache is empty and the next request simply pays full price
once. That keeps the design honest: the cache is an optimization, never a source
of truth, and losing it is never a correctness problem.
