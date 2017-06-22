package utils

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

// Line represents a line of text in context of a byte array
type Line struct {
	Start  int
	End    int
	Source []byte
}

// Text gets the contents of the line
func (line Line) Text() string {
	return string(line.Source[line.Start:line.End])
}

// IsStartOfLine returns true if ind is inside the white space between \n and
// any character
func (line Line) IsStartOfLine(ind int) bool {
	start := line.Start
	if ind < line.Start || ind > line.End {
		return false
	}
	for _, c := range line.Text() {
		// fmt.Printf("Char: %c\n", c)
		if unicode.IsSpace(rune(c)) == true {
			start++
			continue
		}

		// fmt.Printf("Ind: %d, Start: %d\n", ind, start)
		if ind <= start {
			return true
		}

		break
	}
	return false
}

// GetLine takes a byte array and a start position
// returns a line
func GetLine(src []byte, start int) (Line, error) {
	var lineStart, lineEnd int
	l := len(src)

	// Find the end of the line
	for i := start; i < l; i++ {
		if src[i] == '\n' {
			lineEnd = i
			break
		}
	}

	// Find the beginning of the line
	for i := start; i > 0; i-- {
		if src[i] == '\n' {
			lineStart = i
			break
		}
	}

	if lineEnd-lineStart == 0 {
		return Line{}, fmt.Errorf("Empty line")
	}

	return Line{
		Source: src,
		Start:  lineStart,
		End:    lineEnd,
	}, nil
}

// GetCommentBlock takes a byte array and a start position
// and may return a comment block if it finds one
func GetCommentBlock(src []byte, start int) []string {

	pos := start
	linegap := 0
	var comments []string

	line, err := GetLine(src, pos)
	if err != nil {
		return nil
	}

	if line.IsStartOfLine(start) == false {
		return nil
	}

	for {
		line, err := GetLine(src, pos)
		if err != nil {
			break
		}

		trimLine := strings.TrimSpace(line.Text())

		if strings.HasPrefix(trimLine, "#") {
			block := strings.TrimSpace(strings.TrimLeft(trimLine, "#"))
			comments = append([]string{block}, comments...)
			linegap = 0
		}

		if linegap > 0 {
			break
		}

		pos = (line.Start - 1)
		linegap++

		if pos <= 0 {
			break
		}
	}

	return comments
}

// ParseType Convert a schema type to a ast.Type
func ParseType(t string, loc *ast.Location) ast.Type {

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
