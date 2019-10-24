package app

import (
	"fmt"
	"github.com/webitel/engine/model"
	"golang.org/x/net/websocket"
)

type SipProxy struct {
	sock *WebConn
	conn *websocket.Conn
}

func init() {

}

func NewSipProxy(sock *WebConn) *SipProxy {
	p := &SipProxy{
		sock: sock,
	}
	ws, err := websocket.Dial("ws://192.168.177.13:5080", "sip", "http://localhost/")
	if err != nil {
		fmt.Println("Error ", err.Error())
		return nil
	}

	p.conn = ws

	go func() {
		for {
			msg := make([]byte, 8*1024)
			n, err := ws.Read(msg)
			if err != nil {
				panic(err.Error())
			}

			e := model.NewWebSocketEvent("sip")
			e.Add("data", string(msg[:n]))
			fmt.Println("rec:\n", string(msg))
			p.sock.Send <- e
		}
	}()

	return p
}

func (p *SipProxy) Send(data []byte) {
	fmt.Println("send:\n", string(data))
	p.conn.Write(data)
}
