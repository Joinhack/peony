package controllers

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WebSocket struct {
}

func init() {
	go func() {
		for {
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Printf(
				`goroutine number %d
Alloc %d, Sys %d, Frees %d
HeapAlloc %d, HeapSys %d, HeapInuse %d
`, runtime.NumGoroutine(),
				memstats.Alloc, memstats.Sys, memstats.TotalAlloc,
				memstats.HeapAlloc, memstats.HeapSys, memstats.HeapInuse)
			time.Sleep(30 * time.Second)
		}
	}()
}

var hub *pmsg.MsgHub

func init() {
	peony.OnServerInit(func(s *peony.Server) {
		clusterCfg := s.App.GetStringConfig("cluster", "")
		whoami := s.App.GetStringConfig("whoami", "")
		if clusterCfg == "" || whoami == "" {
			panic("please set cluster")
		}
		clusterMap := map[string]string{}
		clusters := strings.Split(clusterCfg, ",")
		//hook log
		pmsg.ERROR = peony.ERROR
		pmsg.WARN = peony.WARN
		pmsg.INFO = peony.INFO
		pmsg.TRACE = peony.TRACE

		for _, v := range clusters {
			kv := strings.Split(v, "->")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == whoami {
				if i, err := strconv.Atoi(whoami); err != nil {
					panic(err)
				} else {
					hub = pmsg.NewMsgHub(i, 1024*1024, kv[1])
				}

			} else {
				clusterMap[kv[0]] = kv[1]
			}
		}
		for k, v := range clusterMap {
			var i int
			var err error
			if i, err = strconv.Atoi(k); err != nil {
				panic(err)
			}
			hub.AddOutgoing(i, v)
		}
		go hub.ListenAndServe()
		//
	})
}

//@Mapper("/echo", method="WS")
func (c *WebSocket) Echo(ws *websocket.Conn) {
	var bs [1024]byte
	for {
		if n, err := ws.Read(bs[:]); err != nil {
			peony.ERROR.Println(err)
			return
		} else {
			peony.INFO.Println("recv info:", string(bs[:n]))
			ws.Write(bs[:n])
		}
	}
}

//@Mapper("/chat", method="WS")
func (c *WebSocket) ChatSocket(ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(3 * time.Second))
	var register RegisterMsg
	var err error
	if err = websocket.JSON.Receive(ws, &register); err != nil {
		peony.ERROR.Println(err)
		return
	}

	if register.Type != 0 || register.Id == 0 {
		ws.Write(LoginAlterJsonBytes)
		peony.ERROR.Println("please login first")
		return
	}
	if register.DevType != 0x1 && register.DevType != 0x2 {
		ws.Write(UnknownDevicesJsonBytes)
		peony.ERROR.Println("unknown devices")
		return
	}

	if _, err = ws.Write(OkJsonBytes); err != nil {
		peony.ERROR.Println(err)
		return
	}

	mutex := &sync.Mutex{}

	client := NewChatClient(ws, register.Id, register.DevType, mutex)
	go client.SendMsgLoop()
	defer func() {
		if err := recover(); err != nil {
			peony.ERROR.Println(err)
		}
		close(client.wchan)
		hub.RemoveClient(client)
	}()
	hub.AddClient(client)
	var msg Msg
	for {
		ws.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			peony.ERROR.Println(err)
			return
		}
		bs, err := json.Marshal(&msg)
		if err != nil {
			peony.ERROR.Println(err)
			ws.Write(ErrorJsonFormatJsonBytes)
			return
		}
		rmsg := &pmsg.DeliverMsg{To: uint64(msg.To), Carry: bs}
		hub.Dispatch(rmsg)
	}
}
