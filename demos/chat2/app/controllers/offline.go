package controllers

import (
	"github.com/joinhack/peony"
	"github.com/garyburd/redigo/redis"
	"encoding/binary"
)

var (
	offlineRedisPool *redis.Pool
)

func SetOfflineRedisPool(pool *redis.Pool) {
	offlineRedisPool = pool
}

type OfflineController struct {	
}

//@Mapper("/offline/msg_pop_front", methods=["POST", "GET"])
func (oc *OfflineController) MsgPopFront(id string) peony.Renderer {
	var rs  = map[string]interface{}{}
	if id == "" {
		rs["code"] = -1
		rs["error"] = "invaild parameter."
		return peony.RenderJson(rs)
	}
	conn := offlineRedisPool.Get()
	defer conn.Close()
	n, err := redis.Int(conn.Do("msg_pop_front", id))
	if err != nil {
		rs["code"] = -1
		rs["error"] = err.Error()
		return peony.RenderJson(rs)
	}
	rs["code"] = 0
	rs["total"] = n
	return peony.RenderJson(rs)
}

//@Mapper("/offline/msg_rows", methods=["POST", "GET"])
func (oc *OfflineController) MsgRows(id string) peony.Renderer {
	var rs  = map[string]interface{}{}
	if id == "" {
		rs["code"] = -1
		rs["error"] = "invaild parameter."
		return peony.RenderJson(rs)
	}
	conn := offlineRedisPool.Get()
	defer conn.Close()
	n, err := redis.Int(conn.Do("msg_rows", id))
	if err != nil {
		rs["code"] = -1
		rs["error"] = err.Error()
		return peony.RenderJson(rs)
	}
	rs["code"] = 0
	rs["total"] = n
	return peony.RenderJson(rs)
}

func  getContent(val []byte) []byte {
	ptr := val
	var content = make([]byte, 0, len(val))
	content = append(content, '[')
	for len(ptr) > 0 {
		l := int(binary.LittleEndian.Uint16(ptr))
		ptr = ptr[2:]
		content = append(content, ptr[:l]...)
		ptr = ptr[l:]
		if len(ptr) > 0 {
			content = append(content, ',')
		}
	}
	content = append(content, ']')
	return content
}


//@Mapper("/offline/msg_front", methods=["POST", "GET"])
func (oc *OfflineController) MsgFront(id string) peony.Renderer {
	var rs  = map[string]interface{}{}
	if id == "" {
		rs["code"] = -1
		rs["error"] = "invaild parameter."
		return peony.RenderJson(rs)
	}
	conn := offlineRedisPool.Get()
	defer conn.Close()
	var err error
	var val []byte
	
	m, e := conn.Do("msg_front", id)
	val, err = redis.Bytes(m, e)
	if err != nil && err != redis.ErrNil {
		rs["code"] = -1
		rs["error"] = err.Error()
		return peony.RenderJson(rs)
	}

	return &peony.TextRenderer{TextSlice: getContent(val), ContentType:"application/json"}
}

