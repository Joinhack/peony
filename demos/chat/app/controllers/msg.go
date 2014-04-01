package controllers

import (
	"encoding/json"
)

var (
	OkJsonBytes, _             = json.Marshal(map[string]interface{}{"code": 0})
	LoginAlterJsonBytes, _     = json.Marshal(map[string]interface{}{"code": -1, "msg": "please login"})
	UnknownDevicesJsonBytes, _ = json.Marshal(map[string]interface{}{"code": -1, "msg": "unknown devices"})

	ErrorJsonFormatJsonBytes, _ = json.Marshal(map[string]interface{}{"code": -1, "msg": "json format error"})
)

type RegisterMsg struct {
	Id      uint64 `json:"id"`
	DevType byte   `json:"devType"`
	Type    byte   `json:"type"`
}

type Msg struct {
	From    uint64 `json:"from"`
	To      uint64 `json:"to"`
	Type    byte   `json:"type"`
	Content string `json:"content"`
}
