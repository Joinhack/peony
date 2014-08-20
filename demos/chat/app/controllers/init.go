package controllers

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/joinhack/peony"
	"github.com/joinhack/pmsg"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	hub *pmsg.MsgHub

	pusher *Pusher

	groupRedisPool *redis.Pool

	tokenRedisPool *redis.Pool

	TRACE *log.Logger

	ERROR *log.Logger

	INFO *log.Logger

	WARN *log.Logger

	hubAddrs = map[int]string{}
)

func hookLog() {
	pmsg.ERROR = peony.ERROR
	pmsg.WARN = peony.WARN
	pmsg.INFO = peony.INFO
	pmsg.TRACE = peony.TRACE

	ERROR = peony.ERROR
	WARN = peony.WARN
	INFO = peony.INFO
	TRACE = peony.TRACE
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}


var statusFmt = `goroutine number %d
Alloc %d, Sys %d, Frees %d
HeapAlloc %d, HeapSys %d, HeapInuse %d
`


func init() {

	go func() {
		for {
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Printf(statusFmt,
				runtime.NumGoroutine(),
				memstats.Alloc, memstats.Sys, memstats.TotalAlloc,
				memstats.HeapAlloc, memstats.HeapSys, memstats.HeapInuse)
			time.Sleep(60 * time.Second)
		}
	}()

	peony.OnServerInit(func(s *peony.Server) {

		pushsvr := s.App.GetStringConfig("push.url", "")
		pushnum := s.App.GetStringConfig("push.num", "")
		groupServer := s.App.GetStringConfig("group.server", "")
		groupPasswd := s.App.GetStringConfig("group.passwd", "")
		rangeVal := s.App.GetIntConfig("range", 1024*1024)
		tokenServer := s.App.GetStringConfig("token.server", "")
		tokenPasswd := s.App.GetStringConfig("token.passwd", "")
		whoami := s.App.GetStringConfig("whoami", "")
		offlineRange := s.App.GetStringConfig("offlineRange", "")
		offlineStorePath := s.App.GetStringConfig("offlineStorePath", "")

		clusterCfg := s.App.GetStringConfig("cluster", "")
		clusterMap := map[string]string{}
		clusters := strings.Split(clusterCfg, ",")
		//hook log
		hookLog()

		if groupServer != "" {
			groupRedisPool = newPool(groupServer, groupPasswd)
			rconn := groupRedisPool.Get()
			if _, err := rconn.Do("ping"); err != nil {
				panic(err)
			}
			rconn.Close()
		}

		if tokenServer != "" {
			tokenRedisPool = newPool(tokenServer, tokenPasswd)
			rconn := tokenRedisPool.Get()
			if _, err := rconn.Do("ping"); err != nil {
				panic(err)
			}
			rconn.Close()
		}

		cfg := &pmsg.MsgHubFileStoreOfflineCenterConfig{}
		cfg.MaxRange = uint32(rangeVal)
		for _, v := range clusters {
			kv := strings.Split(v, "->")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == whoami {
				if i, err := strconv.Atoi(whoami); err != nil {
					panic(err)
				} else {
					cfg.Id = i
					cfg.ServAddr = kv[1]
				}
			} else {
				clusterMap[kv[0]] = kv[1]
			}
		}
		rangeStr := strings.Split(offlineRange, "-")
		if i, err := strconv.Atoi(rangeStr[0]); err != nil {
			panic(err)
		} else {
			cfg.OfflineRangeStart = uint64(i)
		}

		if pushsvr != "" && pushnum != "" {
			var num int
			if i, err := strconv.Atoi(pushnum); err != nil {
				panic(err)
			} else {
				num = i
			}
			pusher = NewPusher(num, pushsvr)
		}

		if i, err := strconv.Atoi(rangeStr[1]); err != nil {
			panic(err)
		} else {
			cfg.OfflineRangeEnd = uint64(i)
		}
		cfg.OfflinePath = offlineStorePath
		hub = pmsg.NewMsgHubWithFileStoreOfflineCenter(cfg)
		for k, v := range clusterMap {
			var i int
			var err error
			if i, err = strconv.Atoi(k); err != nil {
				panic(err)
			}
			hub.AddOtherHub(i, v)
			hubAddrs[i] = v
		}
		hub.AddOfflineMsgFilter(sendNotify)
		go hub.ListenAndServe()
	})
}
