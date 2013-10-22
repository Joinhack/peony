package peony

import ()

type Error struct {
	FileName   string
	SouceLines []string
	Path       string
	Line       int
	Column     int
	Desciption string
}

func (e *Error) Error() string {
	return e.Desciption
}
