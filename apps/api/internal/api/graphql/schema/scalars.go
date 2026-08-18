// Package schema provides GraphQL schema definitions.
package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// JSONScalar is a custom scalar for arbitrary JSON values.
// It strictly validates that serialized and parsed values are valid JSON types.
// (object, array, string, number, boolean, null) and rejects anything else.
var JSONScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSON",
	Description: "Arbitrary JSON value (object, array, string, number, boolean, or null)",
	Serialize: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		if !isValidJSONType(value) {
			panic(fmt.Sprintf("JSON scalar cannot serialize value of type %T", value))
		}
		return value
	},
	ParseValue: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		if !isValidJSONType(value) {
			return nil
		}
		return value
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		return jsonLiteralToGo(valueAST)
	},
})

// DateTimeScalar is a custom scalar for ISO 8601 datetime strings.
// It strictly validates the format on both serialization and parsing.
var DateTimeScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "DateTime",
	Description: "ISO 8601 datetime string (e.g. 2024-01-15T10:30:00Z)",
	Serialize: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		switch v := value.(type) {
		case time.Time:
			if v.IsZero() {
				return nil
			}
			return v.UTC().Format(time.RFC3339Nano)
		case string:
			if v == "" {
				return nil
			}
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				t, err = time.Parse(time.RFC3339Nano, v)
			}
			if err != nil {
				panic(fmt.Sprintf("DateTime scalar cannot serialize non-ISO-8601 string: %s", v))
			}
			return t.UTC().Format(time.RFC3339Nano)
		case *time.Time:
			if v == nil || v.IsZero() {
				return nil
			}
			return v.UTC().Format(time.RFC3339Nano)
		default:
			panic(fmt.Sprintf("DateTime scalar cannot serialize value of type %T", value))
		}
	},
	ParseValue: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		s, ok := value.(string)
		if !ok {
			return nil
		}
		t, err := parseDateTime(s)
		if err != nil {
			return nil
		}
		return t
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		lit, ok := valueAST.(*ast.StringValue)
		if !ok {
			return nil
		}
		t, err := parseDateTime(lit.Value)
		if err != nil {
			return nil
		}
		return t
	},
})

// isValidJSONType reports whether v is a Go type that maps directly to a JSON value.
func isValidJSONType(v interface{}) bool {
	switch v.(type) {
	case nil, bool, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		json.Number, json.RawMessage:
		return true
	case map[string]interface{}:
		return true
	case []interface{}:
		return true
	default:
		return false
	}
}

// jsonLiteralToGo converts a GraphQL AST literal into the corresponding Go value.
// Returns nil for any AST node that does not represent valid JSON.
func jsonLiteralToGo(valueAST ast.Value) interface{} {
	switch v := valueAST.(type) {
	case *ast.IntValue:
		var i int64
		_, err := fmt.Sscanf(v.Value, "%d", &i)
		if err != nil {
			return nil
		}
		return i
	case *ast.FloatValue:
		var f float64
		_, err := fmt.Sscanf(v.Value, "%g", &f)
		if err != nil {
			return nil
		}
		return f
	case *ast.StringValue:
		return v.Value
	case *ast.BooleanValue:
		return v.Value
	case *ast.ListValue:
		values := make([]interface{}, 0, len(v.Values))
		for _, elem := range v.Values {
			values = append(values, jsonLiteralToGo(elem))
		}
		return values
	case *ast.ObjectValue:
		m := make(map[string]interface{})
		for _, field := range v.Fields {
			if field.Name == nil {
				return nil
			}
			m[field.Name.Value] = jsonLiteralToGo(field.Value)
		}
		return m
	case *ast.EnumValue:
		if v.Value == "null" {
			return nil
		}
		return nil
	default:
		return nil
	}
}

// parseDateTime parses an ISO 8601 datetime string, accepting RFC 3339.
// and RFC 3339 with nanosecond precision. Returns an error for invalid formats.
func parseDateTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid ISO 8601 datetime: %s", s)
}
