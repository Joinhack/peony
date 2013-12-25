package controllers

import (
	. "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chat/app/chatroom"
)

//@Mapper("/longpolling/")
type LongPolling struct {
}

func (c LongPolling) Room(user string) Render {
	chatroom.Join(user)
	return AutoRender(map[string]interface{}{"user": user})
}

//@Mapper("/longpolling/room/messages", method="POST")
func (c LongPolling) Say(user, message string) Render {
	chatroom.Say(user, message)
	return nil
}

//@Mapper("/longpolling/room/messages", method="GET")
func (c LongPolling) WaitMessages(lastReceived int) Render {
	subscription := chatroom.Subscribe()
	defer subscription.Cancel()

	// See if anything is new in the archive.
	var events []chatroom.Event
	for _, event := range subscription.Archive {
		if event.Timestamp > lastReceived {
			events = append(events, event)
		}
	}

	// If we found one, grand.
	if len(events) > 0 {
		return AutoRender(events)
	}

	// Else, wait for something new.
	event := <-subscription.New
	return AutoRender([]chatroom.Event{event})
}

//@Mapper("/longpolling/room/leave")
func (c LongPolling) Leave(user string) Render {
	chatroom.Leave(user)
	return NewRedirectRender(Application.Index)
}
