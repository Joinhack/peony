package chat

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

	userRedisPool *redis.Pool

	tokenRedisPool *redis.Pool

	offlineRedisPool *redis.Pool

	TRACE *log.Logger

	ERROR *log.Logger

	INFO *log.Logger

	WARN *log.Logger

	hubAddrs = map[int]string{}
)

func GetOfflineRedisPool() *redis.Pool {
	return offlineRedisPool
}

func GetUserRedisPool() *redis.Pool {
	return userRedisPool
}

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

func Init() {

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
		userServer := s.App.GetStringConfig("user.server", "")
		userPasswd := s.App.GetStringConfig("user.passwd", "")
		rangeVal := s.App.GetIntConfig("range", 1024*1024)
		tokenServer := s.App.GetStringConfig("token.server", "")
		tokenPasswd := s.App.GetStringConfig("token.passwd", "")
		whoamiRaw := s.App.GetStringConfig("whoami", "")
		offlineServer := s.App.GetStringConfig("offline.server", "")
		offlinePasswd := s.App.GetStringConfig("offline.passwd", "")

		clusterCfg := s.App.GetStringConfig("cluster", "")
		clusterMap := map[string]string{}
		clusters := strings.Split(clusterCfg, ",")
		//hook log
		hookLog()

		if offlineServer != "" {
			offlineRedisPool = newPool(offlineServer, offlinePasswd)
			rconn := offlineRedisPool.Get()
			if _, err := rconn.Do("ping"); err != nil {
				panic(err)
			}
			rconn.Close()
		}

		if userServer != "" {
			userRedisPool = newPool(userServer, userPasswd)
			rconn := userRedisPool.Get()
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
		offlineCenter := pmsg.NewRedisClientOfflineCenter(offlineRedisPool)
		
		var serverAddr string
		var whoami int
		for _, v := range clusters {
			kv := strings.Split(v, "->")
			if len(kv) != 2 {
				continue
			}
			if kv[0] == whoamiRaw {
				if i, err := strconv.Atoi(whoamiRaw); err != nil {
					panic(err)
				} else {
					serverAddr = kv[1]
					whoami = i
				}
			} else {
				clusterMap[kv[0]] = kv[1]
			}
		}

		hub = pmsg.NewMsgHubWithOfflineCenter(whoami, uint32(rangeVal), serverAddr, offlineCenter)

		if pushsvr != "" && pushnum != "" {
			var num int
			if i, err := strconv.Atoi(pushnum); err != nil {
				panic(err)
			} else {
				num = i
			}
			pusher = NewPusher(num, pushsvr)
		}

		for k, v := range clusterMap {
			var i int
			var err error
			if i, err = strconv.Atoi(k); err != nil {
				panic(err)
			}
			hub.AddOtherHub(i, v)
			hubAddrs[i] = v
		}
		//hub.AddOfflineMsgFilter(sendNotify)
		go hub.ListenAndServe()
	})
}
