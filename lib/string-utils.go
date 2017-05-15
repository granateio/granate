package lib

import "strings"

type Line struct {
	Text  string
	Start int
	End   int
}

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

func FindCommentBlock(src []byte, start int) string {
	// TODO(nohack) Add support for multiline comments
	// TODO: Make the empty line gap a maximum of 1-2 lines

	pos := start
	gap := 0
	block := ""

	for {
		line := FindLine(src, pos)
		trimLine := strings.TrimSpace(line.Text)

		if strings.HasPrefix(trimLine, "#") {
			block = strings.TrimLeft(trimLine, "# ") + block
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

	return block
}
