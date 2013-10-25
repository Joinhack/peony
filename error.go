package peony

import (
	"fmt"
	"strings"
)

type Error struct {
	Title       string
	FileName    string
	SouceLines  []string
	Path        string
	Line        int
	Column      int
	Description string
}

type ErrorList []*Error

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.FileName, e.Line, e.Column, e.Description)
}

func (el ErrorList) Error() string {
	var r = []string{}
	for _, err := range el {
		r = append(r, err.Error())
	}
	return strings.Join(r, "\n")
}
