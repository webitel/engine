package b2bua

import (
	"context"
	"fmt"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/webitel/engine/b2bua/account"
	"github.com/webitel/engine/b2bua/call"
	"github.com/webitel/engine/b2bua/session"
	"github.com/webitel/engine/b2bua/ua"
	"sync"
)

type AuthInfo struct {
	DomainId    int64
	UserId      int64
	DisplayName string
	Expires     uint32
	account.AuthInfo
}

type Account struct {
	auth      AuthInfo
	recipient sip.SipUri
	register  *ua.Register
	profile   *account.Profile
	ctx       context.Context

	calls []*call.Call

	sess         *session.Session
	mx           sync.RWMutex
	unregisterCh chan struct{}
}

func (b2b *B2B) NewAccount(auth AuthInfo, doRegister account.DoRegister) (*Account, error) {
	var err error
	var uri sip.Uri

	a := &Account{
		auth:      auth,
		recipient: sip.SipUri{},
		register:  nil,
		ctx:       context.TODO(),
	}

	uri, err = parser.ParseUri(fmt.Sprintf("sip:%s@%s", auth.Extension, auth.Realm)) // this acts as an identifier, not connection info
	if err != nil {
		return nil, err
	}
	a.profile = account.NewProfile(auth.DomainId, auth.UserId, uri.Clone(), auth.DisplayName, &a.auth.AuthInfo, auth.Expires, b2b.stack, doRegister)

	a.recipient, err = parser.ParseSipUri(fmt.Sprintf("sip:%s@%s;transport=%s", auth.AuthUser, b2b.host, b2b.transport)) // this is the remote address
	if err != nil {
		return nil, err
	}

	a.register, err = b2b.ua.SendRegister(a.profile, a.recipient, a.profile.Expires, a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Account) setUnregisterChan(ch chan struct{}) {
	a.mx.Lock()
	a.unregisterCh = ch
	a.mx.Unlock()
}

func (a *Account) getUnregisterChan() chan struct{} {
	a.mx.RLock()
	ch := a.unregisterCh
	a.mx.RUnlock()
	return ch
}

func (a *Account) Register() error {
	//a.register.SendRegister(a.auth.Expires)
	return nil
}

func (a *Account) UnRegister() error {
	err := a.register.SendRegister(0)
	a.register.Stop()
	return err
}
