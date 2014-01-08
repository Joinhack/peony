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

// @Mapper(`/static/<re(.*):path>`)
func FileDown(path string) app.Renderer {
	return app.RenderFile(path)
}

// @Mapper(url="/")
func (l *Login) Index(user []*Mail, m *model.User) app.Renderer {
	println(l.xx)
	return app.Render("welcome~!")
}

// @Mapper(ignore=true)
//@Intercept("BEFORE", priority=1)
func (l *Login) Before(c *app.Controller) app.Renderer {
	l.xx = "1000"
	return nil
}

// @Mapper(methods=["WS"])
func Test() app.Renderer {
	return app.RenderText("test")
}

// @Mapper
func Index(s *controllers.S, ss string) app.Renderer {
	return app.Render("haha")
}

func (l *Login) Login(user []*Mail, m *model.User) app.Renderer {
	return app.RenderText("success")
}
