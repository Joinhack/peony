package peony

import (
	"fmt"
)

type Error struct {
	Title       string
	FileName    string
	SourceLines []string
	Path        string
	Line        int
	Column      int
	Description string
}

type sourceLine struct {
	Source  string
	Line    int
	IsError bool
}

func (e *Error) ContextSource() []sourceLine {
	if e.SourceLines == nil {
		return nil
	}
	start := (e.Line - 1) - 5
	if start < 0 {
		start = 0
	}
	end := (e.Line - 1) + 5
	if end > len(e.SourceLines) {
		end = len(e.SourceLines)
	}

	if start > end {
		start = end
	}

	var lines []sourceLine = make([]sourceLine, end-start)
	for i, src := range e.SourceLines[start:end] {
		fileLine := start + i + 1
		lines[i] = sourceLine{src, fileLine, fileLine == e.Line}
	}
	return lines
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.FileName, e.Line, e.Column, e.Description)
}
