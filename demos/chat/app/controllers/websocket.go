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
			time.Sleep(5 * time.Second)
		}
	}()
}

type Msg struct {
	Code    int //0 success, -1 error
	Type    int //0 Login, 1 reply, 2 msg
	From    int64
	To      int64
	Content string
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
func (c WebSocket) Echo(ws *websocket.Conn) {
	var bs [1024]byte
	for {
		if n, err := ws.Read(bs[:]); err != nil {
			peony.ERROR.Println(err)
			return
		} else {
			println(bs[:n])
			ws.Write(bs[:n])
		}

	}
}

//@Mapper("/socket", method="WS")
func (c WebSocket) ChatSocket(ws *websocket.Conn) {
	ws.SetReadDeadline(time.Now().Add(30 * time.Second))
	var msg Msg
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		peony.ERROR.Println(err)
		return
	}
	if msg.Type != 0 {
		peony.ERROR.Println("please login first")
		return
	}
	var ok = &Msg{Code: 0, From: msg.From, To: msg.From, Type: 1}
	var fail = &Msg{Code: -1, From: msg.From, To: msg.From, Type: 1}
	websocket.JSON.Send(ws, ok)
	ws.SetReadDeadline(time.Now().Add(1 * time.Hour))
	id := msg.From

	client := pmsg.NewSimpleClientConn(ws, uint64(id), 1)
	defer func() {
		if err := recover(); err != nil {
			peony.ERROR.Println(err)
		}
		client.CloseWchan()
		hub.RemoveClient(client)
	}()

	if err := hub.AddClient(client); err != nil {
		websocket.JSON.Send(ws, fail)
		peony.ERROR.Println(err)
		return
	}

	for {
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			peony.ERROR.Println(err)
			return
		}
		bs, err := json.Marshal(&msg)
		if err != nil {
			peony.ERROR.Println(err)
			websocket.JSON.Send(ws, fail)
			return
		}
		rmsg := &pmsg.DeliverMsg{To: uint64(msg.To), Carry: bs}
		hub.Dispatch(rmsg)
	}
}
