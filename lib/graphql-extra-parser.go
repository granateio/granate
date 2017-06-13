package lib

import (
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// ParseType Convert a schema type to a ast.Type
func ParseType(t string, loc *ast.Location) ast.Type {
	// fmt.Println("ParseType", t)

	if strings.HasSuffix(t, "!") {
		return &ast.NonNull{
			Kind: kinds.NonNull,
			Loc:  loc,
		}
	}

	if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
		return &ast.List{
			Kind: kinds.List,
			Loc:  loc,
		}
	}

	return &ast.Named{
		Kind: kinds.Named,
		Loc:  loc,
	}

}
