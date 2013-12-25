package controllers

import (
	app "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/login/app/controllers/controllers"
	model "github.com/joinhack/peony/demos/login/app/models"
)

// @Mapper
type Login struct {
	xx string
}

//@Mapper
type Mail struct {
}

// @Mapper("/static/<string:path>")
func FileDown(path string) app.Render {
	return app.NewFileRender(path)
}

// @Mapper(url="/")
func (l *Login) Index(user []*Mail, m *model.User) app.Render {
	println(l.xx)
	return app.AutoRender("welcome~!")
}

// @Mapper(ignore=true)
//@Intercept("BEFORE", priority=1)
func (l *Login) Before(c *app.Controller) app.Render {
	l.xx = "1000"
	return nil
}

// @Mapper(methods=["WS"])
func Test() app.Render {
	return app.NewTextRender("test")
}

// @Mapper
func Index(s *controllers.S, ss string) app.Render {
	return app.AutoRender("haha")
}

func (l *Login) Login(user []*Mail, m *model.User) app.Render {
	return app.NewTextRender("success")
}
