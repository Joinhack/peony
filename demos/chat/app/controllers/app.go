package controllers

import (
	"github.com/joinhack/peony"
)

// @Mapper
type Application struct {
}

//@Mapper("/public/<re(.*):path>")
func Public(path string) peony.Render {
	return peony.NewFileRender("/Volumes/joinhack/work/sources/gopath/src/github.com/joinhack/peony/demos/chat/public/" + path)
}

//@Mapper("/")
func (c Application) Index() peony.Render {
	return peony.AutoRender(map[string]interface{}{})
}

func (c Application) EnterDemo(user, demo string) peony.Render {

	// if c.Validation.HasErrors() {
	// 	c.Flash.Error("Please choose a nick name and the demonstration type.")
	// 	return NewRedirectRender(Application.Index)
	// }

	switch demo {
	case "refresh":
		return peony.NewRedirectRender("/refresh?user=%s", user)
	case "longpolling":
		return peony.NewRedirectRender("/longpolling/room?user=%s", user)
	case "websocket":
		return peony.NewRedirectRender("/websocket/room?user=%s", user)
	}
	return peony.NewRedirectRender(Application.Index)
}
