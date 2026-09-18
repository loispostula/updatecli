package yaml

import (
	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func nodeValue(node ast.Node) string {
	var value string

	if err := goyaml.NodeToValue(node, &value); err == nil {
		return value
	}

	return node.String()
}

// setNodeComment attaches comment as a trailing comment of node.
func setNodeComment(node ast.Node, comment string) error {
	commentGroup := &ast.CommentGroupNode{
		Comments: []*ast.CommentNode{
			{
				Token: &token.Token{
					Type: token.CommentType,
					// Add a space before the comment
					Value: " " + comment,
				},
			},
		},
	}

	return node.SetComment(commentGroup)
}
