package utils

import (
	"bytes"
	"fmt"
	"io"
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
func GetCommentBlock(src []byte, index int) (block []string) {

	pos := index
	linegap := false

	for {
		line, err := GetLine(src, pos)
		if err != nil {
			break
		}

		if line.IsStartOfLine(index) == false && linegap == false {
			return
		}

		text := strings.TrimSpace(line.Text())

		if strings.HasPrefix(text, "#") && linegap == true {
			comment := strings.TrimSpace(strings.TrimLeft(text, "#"))

			// Prepend new data to the comment block
			block = append([]string{comment}, block...)
			linegap = false
		}

		if linegap == true {
			break
		}

		pos = line.Start - 1
		linegap = true

		if pos <= 0 {
			break
		}
	}

	return block
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

// LineCounter counts lines in a string/byte buffer
func LineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// Lifo stack
type Lifo struct {
	Stack []interface{}
}

// Push a new element on the top of the stack
func (lifo *Lifo) Push(elem interface{}) {
	prepend := []interface{}{
		elem,
	}
	lifo.Stack = append(prepend, lifo.Stack...)
}

// Pop an element from the top of the stack
func (lifo *Lifo) Pop() interface{} {
	elem := lifo.Stack[0]
	lifo.Stack = lifo.Stack[1:]

	return elem
}

// OpaqueBytesBuffer is a genereic byte buffer interface
type OpaqueBytesBuffer interface {
	GetBuffer() *bytes.Buffer
}

// SwapBuffer is a writeable buffer where the underlying buffer can be swaped
type SwapBuffer struct {
	buffer OpaqueBytesBuffer
}

func (tb SwapBuffer) Write(b []byte) (int, error) {
	return tb.buffer.GetBuffer().Write(b)
}

// SetBuffer swaps the current buffer with b
func (tb *SwapBuffer) SetBuffer(b OpaqueBytesBuffer) {
	tb.buffer = b
}

// GetBuffer gets get current buffer
func (tb SwapBuffer) GetBuffer() OpaqueBytesBuffer {
	return tb.buffer
}
