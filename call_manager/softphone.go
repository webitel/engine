package call_manager

import (
	"github.com/google/uuid"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"math/rand"
)

type Softphone struct {
	Device           SipRegistrationDeviceInfo
	OnTrack          func(track *webrtc.Track)
	OnInvite         func(inviteMessage SipMessage)
	sipInfo          SIPInfoResponse
	wsConn           *websocket.Conn
	fakeDomain       string
	fakeEmail        string
	fromTag          string
	toTag            string
	callId           string
	cseq             int
	messageListeners map[string]func(string)
	inviteKey        string
}

func NewSoftPhone() *Softphone {
	configureLog()
	softphone := Softphone{}
	softphone.OnInvite = func(inviteMessage SipMessage) {}
	softphone.OnTrack = func(track *webrtc.Track) {}
	softphone.fakeDomain = uuid.New().String() + ".invalid"
	softphone.fakeEmail = uuid.New().String() + "@" + softphone.fakeDomain
	softphone.fromTag = uuid.New().String()
	softphone.toTag = uuid.New().String()
	softphone.callId = uuid.New().String()
	softphone.cseq = rand.Intn(10000) + 1
	softphone.messageListeners = make(map[string]func(string))

	softphone.register()
	return &softphone
}

func (softphone *Softphone) addMessageListener(messageListener func(string)) string {
	key := uuid.New().String()
	softphone.messageListeners[key] = messageListener
	return key
}
func (softphone *Softphone) removeMessageListener(key string) {
	delete(softphone.messageListeners, key)
}

func (softphone *Softphone) request(sipMessage SipMessage, responseHandler func(string) bool) {
	log.Debug("↑↑↑\n", sipMessage.ToString())
	if responseHandler != nil {
		var key string
		key = softphone.addMessageListener(func(message string) {
			done := responseHandler(message)
			if done {
				softphone.removeMessageListener(key)
			}
		})
	}

	_, err := softphone.wsConn.Write([]byte(sipMessage.ToString()))
	if err != nil {
		log.Fatal(err)
	}
}

func (softphone *Softphone) response(message string) {
	log.Debug("↑↑↑\n", message)
	_, err := softphone.wsConn.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}
}
