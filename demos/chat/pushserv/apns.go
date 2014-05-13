package pushserv

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"
)

var (
	SendTimeout        = errors.New("send timeout")
	seq         uint32 = 0
)

type Payload struct {
	Alert interface{} `json:"alert"`
	Badge int         `json:"badge"`
	Sound string      `json:"sound,omitempty"`
}

type Notification interface {
	Command() int
	Bytes() []byte
}

type NotificationV1 struct {
	Notification
	id      uint32
	expire  uint32
	token   []byte
	payload []byte
}

func (n *NotificationV1) Command() int {
	return 1
}

func (n *NotificationV1) Bytes() []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(1))
	binary.Write(buffer, binary.BigEndian, uint32(n.id))
	binary.Write(buffer, binary.BigEndian, uint32(time.Second+time.Duration(n.expire)))
	binary.Write(buffer, binary.BigEndian, uint16(len(n.token)))
	binary.Write(buffer, binary.BigEndian, n.token)
	binary.Write(buffer, binary.BigEndian, uint16(len(n.payload)))
	binary.Write(buffer, binary.BigEndian, n.payload)
	return buffer.Bytes()
}

type Client interface {
	SendRequest(req *Request) error
	Close() error
}

type APNSClient struct {
	*tls.Conn
	cert       tls.Certificate
	gateway    string
	notifyChan chan Notification
}

func NewAPNSClient(gw, certPath, keyPath string) (client *APNSClient, err error) {
	cli := APNSClient{gateway: gw, notifyChan: make(chan Notification)}
	if cli.cert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		return
	}
	if err = cli.dial(); err != nil {
		cli.Close()
		return
	}
	client = &cli
	go client.sendTask()
	return
}

func (cli *APNSClient) sendTask() {
	for {
		var notify Notification
		var ok bool
		var err error
		select {
		case notify, ok = <-cli.notifyChan:
			if !ok {
				return
			}
		}
	SEND:
		if notify != nil {
			if err = cli.Send(notify); err != nil {
				cli.Close()
				if err = cli.dial(); err != nil {
					time.Sleep(20 * time.Millisecond)
					//send again
					goto SEND
				}
			}
		}
		notify = nil
	}
}

func convertRequest2Notification(req *Request) (Notification, error) {
	var err error
	var n = NotificationV1{}
	n.token = req.token
	n.expire = 60 * 60
	n.id = seq
	var p Payload
	p.Alert = req.contents
	p.Badge = 10
	p.Sound = ""
	payload := &struct {
		Aps interface{} `json:"aps"`
	}{p}
	if n.payload, err = json.Marshal(payload); err != nil {
		return nil, err
	}
	seq++
	return &n, nil
}

func (cli *APNSClient) SendRequest(req *Request) error {
	var err error
	var notify Notification
	if notify, err = convertRequest2Notification(req); err != nil {
		return err
	}
	select {
	case cli.notifyChan <- notify:
	case <-time.After(100 * time.Millisecond):
		return SendTimeout
	}
	return nil
}

func (cli *APNSClient) Close() error {
	if cli.Conn != nil {
		return cli.Close()
	}
	return nil
}

func (cli *APNSClient) Send(n Notification) (err error) {
	_, err = cli.Write(n.Bytes())
	return
}

func (cli *APNSClient) dial() (err error) {
	conf := tls.Config{
		Certificates: []tls.Certificate{cli.cert},
	}
	cli.Conn, err = tls.Dial("tcp", cli.gateway, &conf)
	return
}
