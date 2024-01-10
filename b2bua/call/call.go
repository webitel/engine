package call

import (
	"github.com/ghettovoice/gosip/sip"
	"github.com/webitel/engine/b2bua/session"
)

type Call struct {
	Src *session.Session
	//TODO: Add support for forked calls
	Dest     *session.Session
	UserId   int64
	DomainId int64
	WCallId  string
	SockId   string
	Req      *sip.Request
}

func (c *Call) ToString() string {
	return c.Src.Contact() + " => " + c.Dest.Contact()
}
