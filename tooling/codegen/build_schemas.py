#!/usr/bin/env python3
"""Generate the Go wire-type package from CUE schema files.

Reads .cue files from tooling/schema/ and emits Go structs into
apps/api/internal/api/openapi/schemas.go — the package `swag init` reads.

CUE is the single source of truth for the wire DTOs. The TypeScript types are
NOT generated here: they come from the OpenAPI spec via orval (CUE -> Go structs
-> swag init -> openapi3.json -> orval -> generated/schemas). Deriving TS twice
would be a redundant second source; orval already produces a superset (body DTOs
+ *Params + helper types) that the SDK and React Query hooks consume.

CUE file format (deliberately restricted — a small, strict parser, not the full
CUE engine):
    package vyzorix

    #TypeName: {
        fieldName:  type
        optional?:  type
        mapField?:  {[string]: <primitive>}
        nested?:    #OtherType
        list:       [...string] | [...#OtherType]
    }

The parser is strict: any field line it cannot recognize is a hard error, so a
typo or an unsupported CUE construct fails loudly instead of silently dropping
the field from the generated struct.
"""
import os
import re
import sys
import glob

SCHEMA_DIR = os.environ.get("SCHEMA_DIR", "tooling/schema")
# The CUE schemas are the single source of truth for the swaggo DTO package.
# build_schemas.py emits the `openapi` package directly, so `swag init` reads
# generated structs and there is no second hand-maintained copy.
GO_OUT = os.environ.get(
    "GO_OUT", "apps/api/internal/api/openapi/schemas.go"
)
GO_PKG = os.environ.get("GO_PKG", "openapi")

# CUE type → Go type mapping
CUE_TO_GO = {
    "int": "int",
    "int64": "int64",
    "string": "string",
    "bool": "bool",
    "float": "float64",
    "any": "interface{}",
}


def split_types(content):
    """Split a .cue file into (name, body) type blocks with brace-balanced
    bodies — a naive `[^}]*` stops at the `}` inside an open map
    (`{[string]: _}`) and silently truncates those structs."""
    types = []
    for m in re.finditer(r"#(\w+):\s*\{", content):
        name = m.group(1)
        depth = 1
        i = m.end()
        while i < len(content) and depth > 0:
            if content[i] == "{":
                depth += 1
            elif content[i] == "}":
                depth -= 1
            i += 1
        types.append((name, content[m.end():i - 1]))
    return types


def parse_cue_file(filepath):
    """Parse a simplified CUE file into schema definitions."""
    with open(filepath) as f:
        content = f.read()

    # First pass: collect all type definitions (including nested refs)
    type_defs = {}
    for name, body in split_types(content):
        fields = []

        for line in body.strip().split("\n"):
            line = line.strip()
            if not line or line.startswith("//"):
                continue
            # Open map: fieldName?: {[string]: <type>} (map[string]<type>).
            # Value type is `_` (any) or a CUE primitive.
            map_match = re.match(r"(\w+)(\?)?:\s*\{\[string\]:\s*(\w+|_)\}(?:\s*\|\s*\*(.+?))?$", line)
            if map_match:
                fields.append({
                    "name": map_match.group(1),
                    "type": "map",
                    "map_value": map_match.group(3),
                    "optional": map_match.group(2) == "?",
                    "default": map_match.group(4),
                    "is_nested": False,
                })
                continue
            # Any field: fieldName?: _ (interface{} / any / json.RawMessage).
            any_match = re.match(r"(\w+)(\?)?:\s*_(?:\s*\|\s*\*(.+?))?$", line)
            if any_match:
                fields.append({
                    "name": any_match.group(1),
                    "type": "any",
                    "optional": any_match.group(2) == "?",
                    "default": any_match.group(3),
                    "is_nested": False,
                })
                continue
            # Match: fieldName: type | *default  or  fieldName?: type
            # Handles nested refs (#TypeName), arrays ([...T] / [...#T]).
            field_match = re.match(
                r"(\w+)(\?)?:\s*(?:#(\w+)|(\[\.\.\.#?(\w+)\])|(\w+))(?:\s*\|\s*\*(.+?))?$", line
            )
            if field_match:
                fname = field_match.group(1)
                optional = field_match.group(2) == "?"
                # Group 3 = nested type ref (#TypeName)
                # Group 5 = array element type ([...string] or [...#TypeName])
                # Group 6 = primitive type
                nested_ref = field_match.group(3)
                array_type = field_match.group(5)
                primitive_type = field_match.group(6)
                default = field_match.group(7)

                if nested_ref:
                    ftype = nested_ref
                    is_nested = True
                elif array_type:
                    # Array: element type in group 5 (string or #TypeName).
                    # to_go_struct renders [...X] as []X — nested-ref arrays
                    # become []TypeName (not pointers), matching openapi.
                    ftype = f"[...{array_type}]"
                    is_nested = False
                elif primitive_type:
                    ftype = primitive_type
                    is_nested = False
                else:
                    ftype = "any"
                    is_nested = False

                fields.append({
                    "name": fname,
                    "type": ftype,
                    "optional": optional,
                    "default": default,
                    "is_nested": is_nested,
                })
                continue
            # Strict: an unrecognized field line is a hard error, not a silent
            # drop. This is what keeps the restricted CUE subset honest.
            raise ValueError(
                f"{os.path.basename(filepath)}: cannot parse line in #{name}: {line!r}\n"
                "  Supported: `name: type`, `name?: type`, `name?: {[string]: <primitive>}`, "
                "`name?: #Type`, `name: [...type] | [...#Type]`."
            )

        if fields:
            type_defs[name] = fields

    # Second pass: convert to schema list (keep nested refs as type references, don't inline)
    schemas = []
    for name, fields in type_defs.items():
        schemas.append({"name": name, "fields": fields})

    return schemas


def to_go_struct(schema):
    """Convert a schema to a Go struct definition. Fields are emitted in
    fieldalignment-optimal order (pointers/strings/slices before scalars) so
    the generated structs satisfy govet's fieldalignment check without padding."""
    # Resolve Go types first so we can sort by alignment cost.
    resolved = []
    for field in schema["fields"]:
        if field["type"] == "map":
            mv = field.get("map_value", "_")
            go_type = f"map[string]{CUE_TO_GO.get(mv, mv)}" if mv != "_" else "map[string]any"
        elif field["type"] == "any":
            go_type = "any"
        elif field.get("is_nested"):
            go_type = f"*{field['type']}" if field["optional"] else field["type"]
        elif field["type"].startswith("[..."):
            elem = field["type"].replace("[...", "").replace("]", "")
            go_type = f"[]{elem}"
        else:
            go_type = CUE_TO_GO.get(field["type"], field["type"])
            if field["type"] in ("int64",):
                go_type = f"*{go_type}" if field["optional"] else go_type
        resolved.append((field, go_type))

    # fieldalignment: group by field size so wide fields (pointer words: ptr,
    # string, slice, iface) come first, then 8-byte scalars, then narrower
    # ints, then bools — minimizing inter-field padding.
    def size_class(go_type):
        if go_type.startswith("*") or go_type == "string" or go_type == "interface{}":
            return 0  # 1-2 pointer words
        if go_type.startswith("[]"):
            return 1  # slice: 3 pointer words
        if go_type in ("int64", "float64", "int", "uint64"):
            return 2  # 8-byte
        if go_type in ("int32", "float32", "uint32"):
            return 3  # 4-byte
        if go_type in ("int16", "int8", "byte", "bool"):
            return 4  # 1-2 byte
        return 0

    resolved.sort(key=lambda fg: (size_class(fg[1]), fg[0]["name"]))

    lines = [f"type {schema['name']} struct {{"]
    for field, go_type in resolved:
        json_tag = field["name"]
        if field["optional"]:
            json_tag += ",omitempty"
        lines.append(f"\t{pascal(field['name'])} {go_type} `json:\"{json_tag}\"`")
    lines.append("}")
    return "\n".join(lines)


def pascal(name):
    """Wire field name (json tag) → Go field name (PascalCase, preserving
    acronyms the openapi package uses: ID, URL, FCM, MFA, OS, IMEI, HMAC, API).
    Trailing single-letter units like `ms` stay lowercase (RequestTimeoutMs);
    `ids` pluralizes to `IDs` (DeviceIDs)."""
    # Words split on underscores and lower→upper transitions.
    words = re.split(r"[_]|(?<=[a-z])(?=[A-Z])", name)
    words = [w for w in words if w]
    result = []
    for i, word in enumerate(words):
        low = word.lower()
        if word.isupper():
            result.append(word)
        elif low in ACRONYMS:
            result.append(low.upper())
        elif low in LOWER_UNITS:
            result.append(word[0].upper() + low[1:])
        elif low == "ids":
            result.append("IDs")
        else:
            result.append(word[0].upper() + word[1:])
    return "".join(result)


# Acronyms kept all-caps in Go field names.
ACRONYMS = {
    "id", "url", "uri", "fcm", "api", "jwt", "otp", "mfa", "os", "imei",
    "http", "https", "tcp", "udp", "ssl", "tls", "dns", "ip", "mac",
    "db", "sql", "ws", "totp", "csrf", "uuid", "sdk", "ota",
}
# Units that must stay lowercase in Go field names (RequestTimeoutMs, not MS).
LOWER_UNITS = {"ms"}


def main():
    cue_files = sorted(glob.glob(os.path.join(SCHEMA_DIR, "*.cue")))
    if not cue_files:
        print(f"No .cue files found in {SCHEMA_DIR}")
        return

    all_schemas = []
    for cue_file in cue_files:
        schemas = parse_cue_file(cue_file)
        all_schemas.extend(schemas)
        print(f"  {os.path.basename(cue_file)}: {len(schemas)} schemas")

    # Generate Go
    go_dir = os.path.dirname(GO_OUT)
    os.makedirs(go_dir, exist_ok=True)
    with open(GO_OUT, "w") as f:
        f.write("// Code generated by tooling/codegen/build_schemas.py from .cue files.\n")
        f.write("// DO NOT EDIT MANUALLY.\n")
        f.write("//\n")
        f.write("// This package holds the request/response structs that back the\n")
        f.write("// swaggo annotations. swag init parses these into swagger.json\n")
        f.write("// definitions, which build_openapi3.py lifts into OpenAPI3\n")
        f.write("// components.schemas so orval generates real TypeScript types.\n\n")
        f.write(f"package {GO_PKG}\n\n")
        for schema in all_schemas:
            f.write(to_go_struct(schema) + "\n\n")

    print(f"\nGenerated {len(all_schemas)} schemas -> {GO_OUT}")
    print("TypeScript comes from the spec via orval (not generated here).")


if __name__ == "__main__":
    main()
