// Package schema provides GraphQL schema definitions.
package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// JSONScalar is a custom scalar for arbitrary JSON values.
var JSONScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSON",
	Description: "Arbitrary JSON value",
	Serialize: func(value interface{}) interface{} {
		return value
	},
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		return valueAST
	},
})

// DateTimeScalar is a custom scalar for ISO 8601 datetime.
var DateTimeScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "DateTime",
	Description: "ISO 8601 datetime string",
	Serialize: func(value interface{}) interface{} {
		if value == nil {
			return nil
		}

		return value
	},
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		return valueAST
	},
})
