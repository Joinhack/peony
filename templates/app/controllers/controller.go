package controllers

import (
	"github.com/joinhack/peony"
)

//@Mapper("/")
func Index() peony.Renderer {
	return peony.Render()
}
