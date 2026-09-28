package parser_test

import (
	"testing"

	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/file"
	"github.com/antonmedv/expr/parser"
	"github.com/stretchr/testify/require"
)

func TestNamespacedIdentifiers(t *testing.T) {
	for _, input := range []string{
		"data:age >= 18", "system:user.ID == 1", "var:items[0].name",
		"Custom:field and enabled", "data:index",
		"flag ? (data:first) : data:second", "items[(data:index)]",
		"{a: data:value, 'url:x': var:name}",
		"flag?a:b", "flag?true:false", "items[start:end]", "flag ? foo(data:age) : b",
		"flag ? len(data:items) : b", "flag ? items[start:data:end] : b",
		"flag ? [data:age] : b", "flag ? {a:data:age} : b",
		`"https://example.com/a:b"`, `"escaped \" : value"`,
	} {
		t.Run(input, func(t *testing.T) {
			_, err := parser.Parse(input)
			require.NoError(t, err)
		})
	}

	for _, test := range []struct {
		input      string
		want       ast.Node
		wantColumn int
	}{
		{"data:age", &ast.IdentifierNode{Value: "data:age"}, 0},
		{"system:user.ID", &ast.PropertyNode{Node: &ast.IdentifierNode{Value: "system:user"}, Property: "ID"}, 12},
		{"flag?a:b", &ast.ConditionalNode{Cond: &ast.IdentifierNode{Value: "flag"}, Exp1: &ast.IdentifierNode{Value: "a"}, Exp2: &ast.IdentifierNode{Value: "b"}}, 0},
		{"items[start:end]", &ast.SliceNode{Node: &ast.IdentifierNode{Value: "items"}, From: &ast.IdentifierNode{Value: "start"}, To: &ast.IdentifierNode{Value: "end"}}, 5},
		{"flag ? (data:first) : data:second", &ast.ConditionalNode{Cond: &ast.IdentifierNode{Value: "flag"}, Exp1: &ast.IdentifierNode{Value: "data:first"}, Exp2: &ast.IdentifierNode{Value: "data:second"}}, 0},
	} {
		t.Run("ast "+test.input, func(t *testing.T) {
			tree, err := parser.Parse(test.input)
			require.NoError(t, err)
			require.Equal(t, test.input, tree.Source.Content())
			require.Equal(t, ast.Dump(test.want), ast.Dump(tree.Node))
			require.Equal(t, test.wantColumn, tree.Node.Location().Column)
		})
	}
}

func TestMalformedNamespace(t *testing.T) {
	for _, input := range []string{
		"data:", "data:42", "data :age", "data: age", "data::name",
		"a_b:name", "@data:name", "$data:name", "é:name",
		"true:value", "false:value", "nil:value",
	} {
		t.Run(input, func(t *testing.T) {
			_, err := parser.Parse(input)
			require.Error(t, err)
			require.IsType(t, &file.Error{}, err)
			require.Equal(t, 1, err.(*file.Error).Location.Line)
			if input == "data:" {
				require.Equal(t, 4, err.(*file.Error).Location.Column)
			}
		})
	}
}
