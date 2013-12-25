package controllers

import (
	"github.com/joinhack/peony"
)

// @Mapper
type Application struct {
}

//@Mapper("/public/<re(.*):path>")
func Public(path string) peony.Renderer {
	return peony.RenderFile("/Volumes/joinhack/work/sources/gopath/src/github.com/joinhack/peony/demos/chat/public/" + path)
}

//@Mapper("/")
func (c Application) Index() peony.Renderer {
	return peony.Render(map[string]interface{}{})
}

func (c Application) EnterDemo(user, demo string) peony.Renderer {

	// if c.Validation.HasErrors() {
	// 	c.Flash.Error("Please choose a nick name and the demonstration type.")
	// 	return NewRedirectRender(Application.Index)
	// }

	switch demo {
	case "refresh":
		return peony.Redirect("/refresh?user=%s", user)
	case "longpolling":
		return peony.Redirect("/longpolling/room?user=%s", user)
	case "websocket":
		return peony.Redirect("/websocket/room?user=%s", user)
	}
	return peony.Redirect(Application.Index)
}
