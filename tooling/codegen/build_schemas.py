#!/usr/bin/env python3
"""Generate Go structs and TypeScript interfaces from CUE schema files.

Reads .cue files from tooling/schema/ and emits:
  - Go structs into apps/api/internal/api/openapi/zz_generated_schemas.go
  - TypeScript interfaces into packages/API_Client/src/generated/schema/

The CUE files are the single source of truth for configuration schemas.
Handlers return the generated Go structs directly (not gin.H), so the
wire format always matches the schema. The OpenAPI spec is then generated
from the Go structs via swaggo, and orval generates TS types from the spec.

CUE file format (simplified — we parse, we don't need the full CUE engine):
    package vyzorix

    #TypeName: {
        fieldName:  type | *default
        optional?: type
    }
"""
import os
import re
import glob

SCHEMA_DIR = "tooling/schema"
GO_OUT = "apps/api/internal/api/schema/zz_generated_schemas.go"
TS_OUT = "packages/API_Client/src/generated/schema"

# CUE type → Go type mapping
CUE_TO_GO = {
    "int": "int",
    "int64": "int64",
    "string": "string",
    "bool": "bool",
    "float": "float64",
    "any": "interface{}",
}

# CUE type → TypeScript type mapping
CUE_TO_TS = {
    "int": "number",
    "int64": "number",
    "string": "string",
    "bool": "boolean",
    "float": "number",
    "any": "unknown",
}


def parse_cue_file(filepath):
    """Parse a simplified CUE file into schema definitions."""
    with open(filepath) as f:
        content = f.read()

    # First pass: collect all type definitions (including nested refs)
    type_defs = {}
    for match in re.finditer(r"#(\w+):\s*\{([^}]*)\}", content, re.DOTALL):
        name = match.group(1)
        body = match.group(2)
        fields = []

        for line in body.strip().split("\n"):
            line = line.strip()
            if not line or line.startswith("//"):
                continue
            # Match: fieldName: type | *default  or  fieldName?: type
            # Handles nested refs (#TypeName), arrays of primitives
            # ([...string]) and arrays of nested refs ([...#TypeName]).
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
        if field.get("is_nested"):
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


def to_ts_interface(schema):
    """Convert a schema to a TypeScript interface."""
    lines = [f"export interface {schema['name']} {{"]
    for field in schema["fields"]:
        # Handle nested type refs
        if field.get("is_nested"):
            ts_type = field["type"]
        elif field["type"].startswith("[..."):
            elem = field["type"].replace("[...", "").replace("]", "")
            ts_type = f"{elem}[]"
        else:
            ts_type = CUE_TO_TS.get(field["type"], field["type"])
        optional = "?" if field["optional"] else ""
        lines.append(f"\t{field['name']}{optional}: {ts_type};")
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
        f.write("// DO NOT EDIT MANUALLY.\n\n")
        f.write("package schema\n\n")
        for schema in all_schemas:
            f.write(to_go_struct(schema) + "\n\n")

    # Generate TypeScript
    os.makedirs(TS_OUT, exist_ok=True)
    with open(os.path.join(TS_OUT, "index.ts"), "w") as f:
        f.write("// Code generated by tooling/codegen/build_schemas.py from .cue files.\n")
        f.write("// DO NOT EDIT MANUALLY.\n\n")
        for schema in all_schemas:
            f.write(to_ts_interface(schema) + "\n\n")

    print(f"\nGenerated {len(all_schemas)} schemas:")
    print(f"  Go:       {GO_OUT}")
    print(f"  TypeScript: {TS_OUT}/index.ts")


if __name__ == "__main__":
    main()
