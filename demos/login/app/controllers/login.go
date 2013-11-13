package controllers

import (
	app "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/login/app/controllers/controllers"
	model "github.com/joinhack/peony/demos/login/app/models"
)

type Login struct {
}

type Mail struct {
}

// @Mapper("/index")
func (l *Login) Index(user []*Mail, m *model.User) app.Render {
	panic(11)
	return app.NewTextRender("")
}

// @Mapper("/test")
func Index(s *controllers.S, ss string) app.Render {
	return app.NewTextRender("")
}
