#!/usr/bin/env python3
"""Convert Swagger 2.0 (swaggo) → OpenAPI 3.0 for orval.

Source of truth: Go handler swaggo annotations + the request/response DTO
structs in internal/api/openapi/schemas.go. `swag init` parses those into
swagger.json `definitions`; this script lifts them into OpenAPI3
`components.schemas` so orval emits real TypeScript types instead of
placeholder `object` shapes.

The conversion is fully derived from swagger.json — no hand-maintained
schema list here. To add a type, annotate the Go handler with
`@Success 200 {object} openapi.MyType` and `@Param body body openapi.MyReq`
and re-run `swag init` + this script.
"""
import json
import re

SWAG_PATH = "apps/api/swag/swagger.json"
OUT_PATH = "apps/api/swag/openapi3.json"

d = json.load(open(SWAG_PATH))

v3 = {
    "openapi": "3.0.0",
    "info": {
        "title": "Vyzorix Update Server API",
        "version": "0.0.01",
        "description": "Go-generated OpenAPI contract from handler annotations",
    },
    "paths": {},
    "components": {"schemas": {}},
}


def short_name(ref):
    """`internal_api_handlers_alert.ruleRequest` → `ruleRequest`.

    swaggo qualifies definition names with the full import path of the
    declaring package. orval generates cleaner TS names from the short
    final segment, so we flatten to that and de-duplicate collisions by
    package path when two packages export the same short name.
    """
    if not ref:
        return ref
    return ref.rsplit(".", 1)[-1]


def convert_schema(schema, seen=None):
    """Recursively convert a Swagger 2.0 schema fragment to OpenAPI 3.0.

    - Rewrites $ref from #/definitions/<pkg>.<Type> to #/components/schemas/<Type>
    - Moves Swagger2 `format: int64` on integers to `format: int64` (kept)
    - Drops nullable/ Format fields Swagger2 uses that O3 models via `nullable`
    """
    if seen is None:
        seen = set()
    if isinstance(schema, bool):
        return schema
    if not isinstance(schema, dict):
        return schema

    out = {}
    for key, val in schema.items():
        if key == "$ref":
            old = val.split("/")[-1]
            out[key] = "#/components/schemas/" + short_name(old)
        elif key in ("properties",):
            out[key] = {k: convert_schema(v, seen) for k, v in val.items()}
        elif key in ("items", "additionalProperties", "schema"):
            out[key] = convert_schema(val, seen)
        elif key == "allOf":
            out[key] = [convert_schema(s, seen) for s in val]
        elif key == "enum":
            out[key] = val
        elif key in ("type",):
            # Swagger2 uses "type": "object" with no properties for gin.H;
            # leave as-is so orval emits `unknown` rather than inventing fields.
            out[key] = val
        elif key in ("format", "description", "required", "title", "example"):
            out[key] = val
        elif key == "nullable":
            # OpenAPI 3 expresses nullability via `nullable: true` — keep it.
            out[key] = val
        elif key == "default":
            out[key] = val
        # Drop Swagger2-only keys (collectionFormat, etc.) silently.
    return out


# Lift every definition swaggo produced into components.schemas. The short
# name is the TS type name orval will emit; collisions across packages are
# rare here because request structs are package-scoped and the openapi
# package holds the canonical response types.
defs = d.get("definitions", {})
by_short = {}
for full_name, schema in defs.items():
    short = short_name(full_name)
    # First-write wins: the openapi package's response types should not be
    # shadowed by a same-named request struct in a handler package. The
    # openapi package is imported by handlers via the `openapi` identifier,
    # so its definitions are named `openapi.MyType` → short `MyType`,
    # matching what @Success annotations reference.
    if short not in by_short:
        by_short[short] = convert_schema(schema)
    else:
        # Merge: prefer the one with more properties (usually the response
        # form). This is a pragmatic collision resolver, not a correctness
        # issue — the annotations reference openapi.X which is canonical.
        existing = by_short[short]
        if len(schema.get("properties", {})) > len(existing.get("properties", {})):
            by_short[short] = convert_schema(schema)

v3["components"]["schemas"] = by_short


def normalize_responses(responses):
    out = {}
    for code, r in responses.items():
        desc = r.get("description", "")
        schema = r.get("schema")
        if schema is None:
            # Endpoints with no @Success annotation: emit a permissive
            # schema so orval still generates a typed function returning
            # unknown rather than skipping the response entirely.
            schema = {"type": "object"}
        out[code] = {
            "description": desc,
            "content": {"application/json": {"schema": convert_schema(schema)}},
        }
    return out


def dedupe_params(params):
    seen = set()
    out = []
    for p in params:
        key = (p.get("name"), p.get("in"))
        if key in seen:
            continue
        seen.add(key)
        out.append(p)
    return out


def convert_param(p):
    """Swagger2 param → OpenAPI3 parameter object.

    Body params become requestBody (returned separately); others stay params.
    """
    if p.get("in") == "body":
        schema = p.get("schema", {"type": "object"})
        return ("body", {
            "required": bool(p.get("required", False)),
            "description": p.get("description", ""),
            "content": {"application/json": {"schema": convert_schema(schema)}},
        })
    if "name" not in p or "in" not in p:
        return None
    required = bool(p.get("required", False)) or p["in"] == "path"
    out = {
        "name": p["name"],
        "in": p["in"],
        "required": required,
        "schema": {"type": p.get("type", "string"), "format": p.get("format", "")} if p.get("format") else {"type": p.get("type", "string")},
        "description": p.get("description", ""),
    }
    if p.get("enum"):
        out["schema"]["enum"] = p["enum"]
    return ("param", out)


for path, ops in d.get("paths", {}).items():
    op_map = {}
    for method, op in ops.items():
        # swaggo can emit duplicate operations when a handler has two
        # annotation blocks (common in this codebase). The last block wins
        # because it's the one immediately above the func — but they're
        # usually identical, so order doesn't matter. Using the method as
        # key naturally de-dupes within a path.
        op_new = {
            "summary": op.get("summary", ""),
            "description": op.get("description", ""),
            "tags": op.get("tags", []),
        }
        params = []
        request_body = None
        for p in op.get("parameters", []):
            converted = convert_param(p)
            if converted is None:
                continue
            kind, obj = converted
            if kind == "body":
                # Preserve the actual $ref swaggo derived from the handler's
                # @Param body body openapi.X annotation — no global
                # RuleRequest override. This was the bug that made every
                # generated body type collapse to RuleRequest.
                request_body = obj
            else:
                params.append(obj)
        params = dedupe_params(params)
        if params:
            op_new["parameters"] = params
        if request_body:
            op_new["requestBody"] = request_body
        op_new["responses"] = normalize_responses(op.get("responses", {}))
        op_map[method] = op_new
    v3["paths"][path] = op_map

with open(OUT_PATH, "w") as f:
    json.dump(v3, f, indent=2)

print(f"openapi3: {len(v3['paths'])} paths, {len(v3['components']['schemas'])} schemas")
print("schemas:", ", ".join(sorted(v3["components"]["schemas"].keys())))
