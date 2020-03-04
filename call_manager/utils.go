package call_manager

import (
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func generateResponse(username, password, realm, method, uri, nonce string) string {
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	response := md5.Sum([]byte(fmt.Sprintf("%x:%s:%x", ha1, nonce, ha2)))
	return fmt.Sprintf("%x", response)
}

func generateAuthorization(sipInfo SIPInfoResponse, method, nonce string) string {
	return fmt.Sprintf(
		`Digest algorithm=MD5, username="%s", realm="%s", nonce="%s", uri="sip:%s", response="%s"`,
		sipInfo.AuthorizationId, sipInfo.Domain, nonce, sipInfo.Domain,
		generateResponse(sipInfo.AuthorizationId, sipInfo.Password, sipInfo.Domain, method, "sip:"+sipInfo.Domain, nonce),
	)
}

func generateProxyAuthorization(sipInfo SIPInfoResponse, method, targetUser, nonce string) string {
	return fmt.Sprintf(
		`Digest algorithm=MD5, username="%s", realm="%s", nonce="%s", uri="sip:%s@%s", response="%s"`,
		sipInfo.AuthorizationId, sipInfo.Domain, nonce, targetUser, sipInfo.Domain,
		generateResponse(sipInfo.AuthorizationId, sipInfo.Password, sipInfo.Domain, method, "sip:"+targetUser+"@"+sipInfo.Domain, nonce),
	)
}

func branch() string {
	return "z9hG4bK" + uuid.New().String()
}

func configureLog() {
	logLevel := "all"
	if logLevel == "all" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}
