package chat

import (
	"encoding/json"
	"fmt"
	"encoding/binary"
	"errors"
	"github.com/garyburd/redigo/redis"
)

var (
	UserNotExist = errors.New("user not exist.")
	UserExists    = errors.New("user already exists.")
	UserIdSeq    = "UserIdSeq"
)

func getGroupMembers(gid uint32) []uint64 {
	members := make([]uint64, 0, 3)
	if userRedisPool != nil {
		var err error
		var reply []interface{}
		conn := userRedisPool.Get()
		defer conn.Close()
		if reply, err = redis.Values(conn.Do("sdump", fmt.Sprintf("%dFT", gid))); err != nil {
			ERROR.Println(err)
			return members
		}
		if len(reply) == 0 {
			return members
		}
		var l int
		var bs []byte
		if _, err = redis.Scan(reply, &l, &bs); err != nil {
			ERROR.Println(err)
			return members
		}
		for i := 0; i < len(bs); i += l {
			var val uint64
			val = uint64(binary.LittleEndian.Uint16(bs[i:]))
			members = append(members, val)
		}
	}
	return members
}

func GetUserByName(name string) (user *UserInfo, err error) {
	conn := userRedisPool.Get()
	defer conn.Close()
	key := fmt.Sprintf("n_%s", name)
	var val []byte
	if val, err = redis.Bytes(conn.Do("get", key)); err != nil {
		if err == redis.ErrNil {
			err = UserNotExist
			return
		}
		return
	}

	uid := binary.LittleEndian.Uint32(val)
	idkey := fmt.Sprintf("u%d", uid)
	if val, err = redis.Bytes(conn.Do("get", idkey)); err != nil {
		if err == redis.ErrNil {
			err = UserNotExist
			return
		}
		return
	}
	
	var u UserInfo
	if err = json.Unmarshal(val, &u); err != nil {
		return
	}
	user = &u
	return
}

func GetUserById(id uint32) (user *UserInfo, err error) {
	conn := userRedisPool.Get()
	defer conn.Close()
	key := fmt.Sprintf("u%d", id)
	var val []byte
	if val, err = redis.Bytes(conn.Do("get", key)); err != nil {
		if err == redis.ErrNil {
			err = UserNotExist
			return
		}
		return
	}
	
	var u UserInfo
	if err = json.Unmarshal(val, &u); err != nil {
		return
	}
	user = &u
	return
}

func AddUser(user *UserInfo) (rs uint32, err error) {
	conn := userRedisPool.Get()
	defer conn.Close()
	var nkey string = fmt.Sprintf("n_%s",user.Name)
	var exist bool
	var seq int64
	if seq, err = redis.Int64(conn.Do("incr", UserIdSeq)); err != nil {
		return
	}
	if exist, err = redis.Bool(conn.Do("exists", nkey)); err != nil {
		return
	}
	if exist {
		err = UserExists
		return
	}
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(seq))
	if _, err = redis.String(conn.Do("set", nkey, buf)); err != nil {
		return
	}
	rs = uint32(seq)
	user.Id = rs
	key := fmt.Sprintf("u%d", rs)
	var val []byte
	if val, err = json.Marshal(user); err != nil {
		return
	}
	if _, err = redis.String(conn.Do("set", key, val)); err != nil {
		return
	}
	return
}
