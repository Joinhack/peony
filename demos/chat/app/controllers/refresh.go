package controllers

import (
	. "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chat/app/chatroom"
)

//@Mapper
type Refresh struct {
}

//@Mapper("/refresh")
func (c Refresh) Index(user string) Renderer {
	chatroom.Join(user)
	return Redirect(Refresh.Room, user)
}

//@Mapper(method="GET")
func (c Refresh) Room(user string) Renderer {
	subscription := chatroom.Subscribe()
	defer subscription.Cancel()
	events := subscription.Archive
	for i, _ := range events {
		if events[i].User == user {
			events[i].User = "you"
		}
	}
	return Render(map[string]interface{}{"user": user, "events": events})
}

//@Mapper("room",method="POST")
func (c Refresh) Say(user, message string) Renderer {
	chatroom.Say(user, message)
	return Redirect(Refresh.Room, user)
}

//@Mapper("room/leave")
func (c Refresh) Leave(user string) Renderer {
	chatroom.Leave(user)
	return Redirect(Application.Index)
}
