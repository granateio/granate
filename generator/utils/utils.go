package utils

import (
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// Line represents a line of text
type Line struct {
	Text  string
	Start int
	End   int
}

// FindLine takes a byte array and a start position
// returns a line
func FindLine(src []byte, start int) Line {
	var lineStart, lineEnd int
	l := len(src)

	for i := start; i < l; i++ {
		if src[i] == '\n' {
			lineEnd = i
			break
		}
	}

	for i := start; i > 0; i-- {
		if src[i] == '\n' {
			lineStart = i
			break
		}
	}

	return Line{
		Text:  string(src[lineStart:lineEnd]),
		Start: lineStart,
		End:   lineEnd,
	}
}

// FindCommentBlock takes a byte array and a start position
// and may return a comment block if it finds one
func FindCommentBlock(src []byte, start int) []string {
	// TODO(nohack) Add support for multiline comments
	// TODO: Make the empty line gap a maximum of 1-2 lines

	pos := start
	gap := 0
	var comments []string

	for {
		line := FindLine(src, pos)
		trimLine := strings.TrimSpace(line.Text)

		if strings.HasPrefix(trimLine, "# - ") {
			block := strings.TrimLeft(trimLine, "# - ")
			comments = append([]string{block}, comments...)
			// fmt.Println(block)
		} else if len(trimLine) > 0 {
			if gap > 0 {
				break
			}
			gap++
		}
		pos = (line.Start - 1)
		if pos <= 0 {
			break
		}
	}

	return comments
}

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
