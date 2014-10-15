package controllers

import (
	"github.com/joinhack/peony/demos/chat2/app/chat"
	"github.com/joinhack/peony"
)

func init() {
	chat.Init()
	
	peony.OnServerInit(func(s *peony.Server) {
		println(chat.GetOfflineRedisPool())
		SetOfflineRedisPool(chat.GetOfflineRedisPool())
	})
}


