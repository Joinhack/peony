package controllers

import (
	"github.com/joinhack/peony"
	"github.com/joinhack/peony/demos/chat2/app/chat"
)


type UserController struct {
}

//@Mapper("/user/login", methods=["POST", "GET"])
func (uc *UserController) Login(name, passwd string) peony.Renderer {
	var rs = map[string]interface{}{}
	
	return peony.RenderJson(rs)
}


//@Mapper("/user/add", methods=["POST", "GET"])
func (uc *UserController) Add(name, passwd string) peony.Renderer {
	var user  = chat.UserInfo{Name:name, Password:passwd}
	var err error
	var rs = map[string]interface{}{}
	var seq uint32
	if len(name) == 0 || len(passwd) == 0 {
		rs["code"] = -1
		rs["msg"] = "parameter error."
		return peony.RenderJson(rs)
	}
	if seq, err = chat.AddUser(&user); err != nil {
		rs["code"] = -1
		rs["msg"] = err.Error() 
	} else {
		rs["code"] = 0
		rs["userId"] = seq
	}
	return peony.RenderJson(rs)
}