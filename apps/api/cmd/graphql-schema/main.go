// Command graphql-schema emits the code-first GraphQL schema (built by
// internal/api/graphql/schema.BuildSchema) as canonical SDL (schema.graphql)
// — the artifact consumed by the drift gate. graphql-go has no SDL printer and
// its executor does not execute the standard introspection __schema query, so
// this renders SDL directly from the built type system (which is the source of
// truth — the introspection query is derived from it).
//
// Run from apps/api:
//
//	go run ./cmd/graphql-schema -out swag/graphql
//
// The schema is built with a nil resolver — BuildSchema only reads the field
// and type definitions (descriptions, args, types), never invokes a resolver,
// so no services/DB are required.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/resolver"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/schema"
	"github.com/graphql-go/graphql"
)

// doc renders a description as a GraphQL block-string doc comment.
func doc(sb *strings.Builder, desc, indent string) {
	if desc == "" {
		return
	}
	sb.WriteString(indent)
	sb.WriteString(`"""`)
	sb.WriteString(desc)
	sb.WriteString(`"""`)
	sb.WriteString("\n")
}

// renderType formats a type reference as SDL (e.g. String!, [Device!]!).
func renderType(t graphql.Type) string {
	switch tt := t.(type) {
	case *graphql.NonNull:
		return renderType(tt.OfType) + "!"
	case *graphql.List:
		return "[" + renderType(tt.OfType) + "]"
	}
	return t.Name()
}

func renderSchema(sch graphql.Schema) string {
	var sb strings.Builder

	// schema {} block — only when a root type name differs from the default.
	var rootLines []string
	if q := sch.QueryType(); q != nil && q.Name() != "Query" {
		rootLines = append(rootLines, "  query: "+q.Name())
	}
	if m := sch.MutationType(); m != nil && m.Name() != "Mutation" {
		rootLines = append(rootLines, "  mutation: "+m.Name())
	}
	if s := sch.SubscriptionType(); s != nil {
		rootLines = append(rootLines, "  subscription: "+s.Name())
	}
	if len(rootLines) > 0 {
		sb.WriteString("schema {\n")
		sb.WriteString(strings.Join(rootLines, "\n"))
		sb.WriteString("\n}\n\n")
	}

	types := sch.TypeMap()
	names := make([]string, 0, len(types))
	for name := range types {
		if !strings.HasPrefix(name, "__") {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		renderTypeDef(&sb, types[name])
	}
	return sb.String()
}

// renderTypeDef renders a single named type definition as SDL.
func renderTypeDef(sb *strings.Builder, t graphql.Type) {
	switch tt := t.(type) {
	case *graphql.Scalar:
		doc(sb, tt.Description(), "")
		sb.WriteString("scalar " + tt.Name() + "\n\n")
	case *graphql.Enum:
		doc(sb, tt.Description(), "")
		sb.WriteString("enum " + tt.Name() + " {\n")
		values := tt.Values()
		sort.Slice(values, func(i, j int) bool { return values[i].Name < values[j].Name })
		for _, v := range values {
			doc(sb, v.Description, "  ")
			sb.WriteString("  " + v.Name + "\n")
		}
		sb.WriteString("}\n\n")
	case *graphql.InputObject:
		doc(sb, tt.Description(), "")
		sb.WriteString("input " + tt.Name() + " {\n")
		fields := tt.Fields()
		names := make([]string, 0, len(fields))
		for name := range fields {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			f := fields[name]
			doc(sb, f.Description(), "  ")
			sb.WriteString("  " + f.Name() + ": " + renderType(f.Type) + "\n")
		}
		sb.WriteString("}\n\n")
	case *graphql.Object:
		renderObject(sb, tt)
	case *graphql.Interface:
		renderInterface(sb, tt)
	case *graphql.Union:
		doc(sb, tt.Description(), "")
		names := make([]string, 0, len(tt.Types()))
		for _, pt := range tt.Types() {
			names = append(names, pt.Name())
		}
		sb.WriteString("union " + tt.Name() + " = " + strings.Join(names, " | ") + "\n\n")
	}
}

// sortedFieldNames returns field names in deterministic (sorted) order —
// graphql-go stores fields in a map, so iteration order is random.
func sortedFieldNames(fields graphql.FieldDefinitionMap) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// renderObject renders an object type (with optional implements clause).
func renderObject(sb *strings.Builder, tt *graphql.Object) {
	doc(sb, tt.Description(), "")
	ifaceClause := ""
	if ifaces := tt.Interfaces(); len(ifaces) > 0 {
		inames := make([]string, 0, len(ifaces))
		for _, i := range ifaces {
			inames = append(inames, i.Name())
		}
		ifaceClause = " implements " + strings.Join(inames, " & ")
	}
	sb.WriteString("type " + tt.Name() + ifaceClause + " {\n")
	fields := tt.Fields()
	for _, name := range sortedFieldNames(fields) {
		f := fields[name]
		doc(sb, f.Description, "  ")
		sb.WriteString("  " + f.Name + renderArgs(f.Args) + ": " + renderType(f.Type) + "\n")
	}
	sb.WriteString("}\n\n")
}

// renderInterface renders an interface type.
func renderInterface(sb *strings.Builder, tt *graphql.Interface) {
	doc(sb, tt.Description(), "")
	sb.WriteString("interface " + tt.Name() + " {\n")
	fields := tt.Fields()
	for _, name := range sortedFieldNames(fields) {
		f := fields[name]
		doc(sb, f.Description, "  ")
		sb.WriteString("  " + f.Name + renderArgs(f.Args) + ": " + renderType(f.Type) + "\n")
	}
	sb.WriteString("}\n\n")
}

// renderArgs formats a field's argument list as SDL: (arg: Type, ...).
func renderArgs(args []*graphql.Argument) string {
	if len(args) == 0 {
		return ""
	}
	names := make([]string, 0, len(args))
	for _, a := range args {
		names = append(names, a.Name())
	}
	sort.Strings(names)
	byName := make(map[string]*graphql.Argument, len(args))
	for _, a := range args {
		byName[a.Name()] = a
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+": "+renderType(byName[name].Type))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func main() {
	outDir := flag.String("out", "swag/graphql", "output directory for schema.graphql")
	flag.Parse()

	// BuildSchema never invokes resolvers — it only reads field/type defs.
	sch, err := schema.BuildSchema(&resolver.Resolver{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "build schema:", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}

	sdl := renderSchema(sch)
	if err := os.WriteFile(filepath.Join(*outDir, "schema.graphql"), []byte(sdl), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write schema.graphql:", err)
		os.Exit(1)
	}

	n := 0
	for name := range sch.TypeMap() {
		if !strings.HasPrefix(name, "__") {
			n++
		}
	}
	fmt.Printf("graphql-schema: %d types → %s/schema.graphql\n", n, *outDir)
}
