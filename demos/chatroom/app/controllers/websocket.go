package controllers

import (
	"code.google.com/p/go.net/websocket"
	. "github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chat/app/chatroom"
)

type WebSocket struct {
}

//@Mapper("/websocket/room")
func (c WebSocket) Room(user string) Renderer {

	return Render(map[string]interface{}{"user": user})
}

//@Mapper("/websocket/room/socket", method="WS")
func (c WebSocket) RoomSocket(user string, ws *websocket.Conn) {
	// Join the room.
	subscription := chatroom.Subscribe()
	defer subscription.Cancel()

	chatroom.Join(user)
	defer chatroom.Leave(user)

	// Send down the archive.
	for _, event := range subscription.Archive {
		if websocket.JSON.Send(ws, &event) != nil {
			// They disconnected
			return
		}
	}

	// In order to select between websocket messages and subscription events, we
	// need to stuff websocket events into a channel.
	newMessages := make(chan string)
	go func() {
		var msg string
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				close(newMessages)
				return
			}
			newMessages <- msg
		}
	}()

	// Now listen for new events from either the websocket or the chatroom.
	for {
		select {
		case event := <-subscription.New:
			if websocket.JSON.Send(ws, &event) != nil {
				// They disconnected.
				return
			}
		case msg, ok := <-newMessages:
			// If the channel is closed, they disconnected.
			if !ok {
				return
			}

			// Otherwise, say something.
			chatroom.Say(user, msg)
		}
	}
	return
}
