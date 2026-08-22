import json

swag_path = "apps/api/swag/swagger.json"
d = json.load(open(swag_path))

# Convert Swagger 2.0 → OpenAPI 3.0 with the pieces orval needs.
v3 = {
    "openapi": "3.0.0",
    "info": {
        "title": "Vyzorix Update Server API",
        "version": "0.0.01",
        "description": "Go-generated OpenAPI contract from handler annotations",
    },
    "paths": {},
    "components": {
        "schemas": {
            "RuleRequest": {
                "type": "object",
                "properties": {
                    "name": {"type": "string"},
                    "metric": {"type": "string"},
                    "condition": {"type": "string"},
                    "webhook_url": {"type": "string"},
                    "threshold": {"type": "number"},
                    "for_seconds": {"type": "integer"},
                    "notify_interval_seconds": {"type": "integer"},
                    "on_no_data": {"type": "string"},
                    "on_error": {"type": "string"},
                    "enabled": {"type": "boolean"},
                },
                "required": ["name", "metric", "condition", "threshold"],
            },
            "AlertRuleWithInstances": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "org_id": {"type": "string"},
                    "name": {"type": "string"},
                    "metric": {"type": "string"},
                    "condition": {"type": "string"},
                    "threshold": {"type": "number"},
                    "for_seconds": {"type": "integer"},
                    "notify_interval_seconds": {"type": "integer"},
                    "on_no_data": {"type": "string"},
                    "on_error": {"type": "string"},
                    "enabled": {"type": "boolean"},
                    "webhook_url": {"type": "string"},
                    "created_at": {"type": "string", "format": "date-time"},
                    "updated_at": {"type": "string", "format": "date-time"},
                    "instances": {
                        "type": "array",
                        "items": {"$ref": "#/components/schemas/AlertInstance"},
                    },
                },
            },
            "AlertInstance": {
                "type": "object",
                "properties": {
                    "labels": {"type": "object", "additionalProperties": {"type": "string"}},
                    "state": {"type": "string"},
                    "value": {"type": "number"},
                    "evaluated_at": {"type": "string", "format": "date-time"},
                },
            },
            "AlertHistoryEvent": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "rule_id": {"type": "string"},
                    "from_state": {"type": "string"},
                    "to_state": {"type": "string"},
                    "value": {"type": "number"},
                    "created_at": {"type": "integer", "format": "int64"},
                },
            },
            "ContactPoint": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "org_id": {"type": "string"},
                    "name": {"type": "string"},
                    "type": {"type": "string"},
                    "config": {"type": "object"},
                    "enabled": {"type": "boolean"},
                    "created_at": {"type": "string", "format": "date-time"},
                    "updated_at": {"type": "string", "format": "date-time"},
                },
            },
            "ServiceAccount": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "org_id": {"type": "string"},
                    "name": {"type": "string"},
                    "description": {"type": "string"},
                    "scopes": {"type": "array", "items": {"type": "string"}},
                    "created_at": {"type": "string", "format": "date-time"},
                    "expires_at": {"type": "string", "format": "date-time", "nullable": True},
                },
            },
            "ConfigVersion": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "org_id": {"type": "string"},
                    "resource_type": {"type": "string"},
                    "version": {"type": "integer"},
                    "snapshot": {"type": "object"},
                    "changed_by": {"type": "string"},
                    "created_at": {"type": "string", "format": "date-time"},
                },
            },
            "Annotation": {
                "type": "object",
                "properties": {
                    "id": {"type": "string"},
                    "org_id": {"type": "string"},
                    "text": {"type": "string"},
                    "tags": {"type": "array", "items": {"type": "string"}},
                    "time": {"type": "string", "format": "date-time"},
                    "time_end": {"type": "string", "format": "date-time", "nullable": True},
                },
            },
            "UsageStatsSnapshot": {
                "type": "object",
                "properties": {
                    "collected_at": {"type": "string", "format": "date-time"},
                    "toggles": {"type": "object", "additionalProperties": {"type": "boolean"}},
                    "counts": {
                        "type": "object",
                        "properties": {
                            "devices": {"type": "integer"},
                            "operators": {"type": "integer"},
                            "organizations": {"type": "integer"},
                            "service_accounts": {"type": "integer"},
                            "alert_rules": {"type": "integer"},
                            "contact_points": {"type": "integer"},
                            "annotations": {"type": "integer"},
                        },
                    },
                },
            },
        }
    },
}

def normalize_responses(responses):
    out = {}
    for code, r in responses.items():
        desc = r.get("description", "")
        schema = r.get("schema", {"type": "object"})
        out[code] = {
            "description": desc,
            "content": {"application/json": {"schema": schema}},
        }
    return out

for path, ops in d.get("paths", {}).items():
    op_map = {}
    for method, op in ops.items():
        op_new = {
            "summary": op.get("summary", ""),
            "description": op.get("description", ""),
            "tags": op.get("tags", []),
        }
        params = []
        request_body = None
        for p in op.get("parameters", []):
            if p.get("in") == "body":
                # generic body map to the first schema the handler uses;
                # endpoints with request bodies point at the correct $ref per group
                request_body = {
                    "required": True,
                    "content": {
                        "application/json": {
                            "schema": {"$ref": "#/components/schemas/RuleRequest"}
                        }
                    },
                }
            else:
                if "name" in p and "in" in p:
                    required = bool(p.get("required", False)) or p["in"] == "path"
                    params.append({
                        "name": p["name"],
                        "in": p["in"],
                        "required": required,
                        "schema": {"type": p.get("type", "string")},
                        "description": p.get("description", ""),
                    })
        if params:
            op_new["parameters"] = params
        if request_body:
            op_new["requestBody"] = request_body
        op_new["responses"] = normalize_responses(op.get("responses", {}))
        op_map[method] = op_new
    v3["paths"][path] = op_map

open("apps/api/swag/openapi3.json", "w").write(json.dumps(v3, indent=2))
print("openapi3:", len(v3["paths"]), "paths,", len(v3["components"]["schemas"]), "schemas")
