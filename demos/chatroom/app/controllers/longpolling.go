package controllers

import (
	. "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chatroom/app/chatroom"
)

//@Mapper("/longpolling/")
type LongPolling struct {
}

func (c LongPolling) Room(user string) Renderer {
	chatroom.Join(user)
	return Render(map[string]interface{}{"user": user})
}

//@Mapper("/longpolling/room/messages", method="POST")
func (c LongPolling) Say(user, message string) Renderer {
	chatroom.Say(user, message)
	return nil
}

//@Mapper("/longpolling/room/messages", method="GET")
func (c LongPolling) WaitMessages(lastReceived int) Renderer {
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
		return Render(events)
	}

	// Else, wait for something new.
	event := <-subscription.New
	return Render([]chatroom.Event{event})
}

//@Mapper("/longpolling/room/leave")
func (c LongPolling) Leave(user string) Renderer {
	chatroom.Leave(user)
	return Redirect(Application.Index)
}
