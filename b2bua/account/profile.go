package account

import (
	"fmt"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/google/uuid"
	"github.com/webitel/engine/b2bua/stack"
	"github.com/webitel/engine/b2bua/utils"
)

var (
	logger log.Logger
)

func init() {
	logger = utils.NewLogrusLogger(log.DebugLevel, "UserAgent", nil)
}

// AuthInfo .
type AuthInfo struct {
	AuthUser  string
	Realm     string
	Password  string
	Ha1       string
	DomainId  int64
	ContactId int64
	Extension string
}

type DoRegister func() (*AuthInfo, error)

// Profile .
type Profile struct {
	URI           sip.Uri
	DisplayName   string
	AuthInfo      *AuthInfo
	Expires       uint32
	InstanceID    string
	Routes        []sip.Uri
	ContactURI    sip.Uri
	ContactParams map[string]string
	CustomHeaders map[string]string
	DomainId      int64
	UserId        int64
	DoRegister    DoRegister
}

// Contact .
func (p *Profile) Contact() *sip.Address {
	var uri sip.Uri
	if p.ContactURI != nil {
		uri = p.ContactURI
	} else {
		uri = p.URI.Clone()
	}

	contact := &sip.Address{
		Uri:    uri,
		Params: sip.NewParams(),
	}
	if p.InstanceID != "nil" {
		contact.Params.Add("+sip.instance", sip.String{Str: p.InstanceID})
	}

	contact.Params.Add("dc", sip.String{Str: fmt.Sprintf("%d", p.DomainId)})
	contact.Params.Add("uid", sip.String{Str: fmt.Sprintf("%d", p.UserId)})
	//contact.Params.Add("expires", sip.String{Str: fmt.Sprintf("%d", p.Expires)})
	for key, value := range p.ContactParams {
		contact.Params.Add(key, sip.String{Str: value})
	}

	//TODO: Add more necessary parameters.
	//etc: ip:port, transport=udp|tcp, +sip.ice, +sip.instance, +sip.pnsreg,

	return contact
}

// NewProfile .
func NewProfile(
	domainId int64,
	userId int64,
	uri sip.Uri,
	displayName string,
	authInfo *AuthInfo,
	expires uint32,
	stack *stack.SipStack,
	doRegister DoRegister,
) *Profile {
	p := &Profile{
		URI:           uri,
		DisplayName:   displayName,
		AuthInfo:      authInfo,
		Expires:       expires,
		CustomHeaders: make(map[string]string),
		DomainId:      domainId,
		UserId:        userId,
		DoRegister:    doRegister,
	}

	if stack != nil { // populate the Contact field
		var transport string
		if tp, ok := uri.UriParams().Get("transport"); ok {
			transport = tp.String()
		} else {
			transport = "udp"
		}
		addr := stack.GetNetworkInfo(transport)
		uri, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s;transport=%s", p.URI.User(), addr.Addr(), transport))
		if err == nil {
			p.ContactURI = uri
		} else {
			logger.Errorf("Error parsing contact URI: %s", err)
		}
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		logger.Errorf("could not create UUID: %v", err)
	}
	p.InstanceID = fmt.Sprintf(`"<%s>"`, uid.URN())
	return p
}

func (p *Profile) AddCustomHeaders(headers map[string]string) {
	p.CustomHeaders = headers
}

// RegisterState .
type RegisterState struct {
	Account    *Profile
	StatusCode sip.StatusCode
	Reason     string
	Expiration uint32
	Response   sip.Response
	UserData   interface{}
}
