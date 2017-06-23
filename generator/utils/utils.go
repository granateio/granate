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

// Text returns the contents of the line as a string
func (line Line) Text() string {
	return string(line.Source[line.Start:line.End])
}

// IsStartOfLine returns true if index is inside the white space between
// the beginning of the line and any non white space character
func (line Line) IsStartOfLine(index int) bool {
	pos := line.Start
	if index < line.Start || index > line.End {
		return false
	}
	for _, c := range line.Text() {
		if unicode.IsSpace(rune(c)) == true {
			pos++
			continue
		}

		if index <= pos {
			return true
		}

		break
	}
	return false
}

// GetLine takes a byte array and a start position
// returns a line
func GetLine(src []byte, index int) (Line, error) {
	start, end := index, index+1
	l := len(src)

	if index > l {
		return Line{}, fmt.Errorf("index out of range")
	}

	// Find the beginning of the line
	for start > 0 && src[start-1] != '\n' {
		start--
	}

	// Find the end of the line
	for end < l && src[end-1] != '\n' {
		end++
	}

	return Line{
		Source: src,
		Start:  start,
		End:    end,
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
