package call_manager

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"net/url"
	"regexp"
	"strings"
)

func (softphone *Softphone) register() {
	softphone.sipInfo = SIPInfoResponse{
		Username:           "202",
		Password:           "1008",
		AuthorizationId:    "1008",
		Domain:             "demo.webitel.com",
		OutboundProxy:      "dev.webitel.com",
		Transport:          "wss",
		Certificate:        "",
		SwitchBackInterval: 0,
	}
	url := url.URL{Scheme: strings.ToLower(softphone.sipInfo.Transport), Host: softphone.sipInfo.OutboundProxy, Path: "/sip"}

	conn, err := websocket.Dial(url.String(), "sip", "http://localhost/")

	if err != nil {
		log.Fatal(err)
	}
	softphone.wsConn = conn
	go func() {
		for {
			var bytes = make([]byte, 8*1024)
			n, err := softphone.wsConn.Read(bytes)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Println(fmt.Sprintf("recv %d: %s\n\n-----", n, string(bytes)))
			message := string(bytes[:n])
			log.Debug("↓↓↓\n", message)
			for _, ml := range softphone.messageListeners {
				go ml(message)
			}
		}
	}()

	sipMessage := SipMessage{}
	sipMessage.method = "REGISTER"
	sipMessage.address = softphone.sipInfo.Domain
	sipMessage.headers = make(map[string]string)
	sipMessage.headers["Contact"] = fmt.Sprintf("<sip:%s;transport=ws>;expires=600", softphone.fakeEmail)
	sipMessage.headers["Via"] = fmt.Sprintf("SIP/2.0/WSS %s;branch=%s", softphone.fakeDomain, branch())
	sipMessage.headers["From"] = fmt.Sprintf("<sip:%s@%s>;tag=%s", softphone.sipInfo.Username, softphone.sipInfo.Domain, softphone.fromTag)
	sipMessage.headers["To"] = fmt.Sprintf("<sip:%s@%s>", softphone.sipInfo.Username, softphone.sipInfo.Domain)
	sipMessage.addCseq(softphone).addCallId(*softphone).addUserAgent()
	softphone.request(sipMessage, func(message string) bool {
		if strings.Contains(message, "WWW-Authenticate: Digest") {
			authenticateHeader := SipMessage{}.FromString(message).headers["WWW-Authenticate"]
			regex := regexp.MustCompile(`, nonce="(.+?)"`)
			nonce := regex.FindStringSubmatch(authenticateHeader)[1]

			sipMessage.addAuthorization(*softphone, nonce).addCseq(softphone).newViaBranch()
			softphone.request(sipMessage, nil)
			return true
		}
		return false
	})
}
