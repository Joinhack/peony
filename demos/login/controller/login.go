package controller

import (
	"github.com/joinhack/peony"
	model "github.com/joinhack/peony/demos/login/models"
)

type Login struct {
}

type Mail struct {
}

// @Mapper("/index")
func (l *Login) Index(user *model.User) Render {
	return peony.NewTemplateRender()
}

// @Mapper("/")
func Index(a, b string) Render {
	return peony.NewTemplateRender()
}
