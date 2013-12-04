package controllers

import (
	app "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/login/app/controllers/controllers"
	model "github.com/joinhack/peony/demos/login/app/models"
)

type Login struct {
	xx string
}

type Mail struct {
}

func FileDown() app.Render {
	return NewFileRender(".")
}

// @Mapper("/")
func (l *Login) Index(user []*Mail, m *model.User) app.Render {
	println(l.xx)
	return nil
}

// @Intercept(Before,1)
func (l *Login) Before(c *app.Controller) app.Render {
	l.xx = "1000"
	return nil
}

// @Mapper("/test")
func Index(s *controllers.S, ss string) app.Render {
	return app.NewTextRender("")
}
