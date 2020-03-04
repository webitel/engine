package call_manager

import (
	"fmt"
	"strings"
)

func (softphone *Softphone) OpenToInvite() {
	softphone.inviteKey = softphone.addMessageListener(func(message string) {
		if strings.HasPrefix(message, "INVITE sip:") {
			inviteMessage := SipMessage{}.FromString(message)

			dict := map[string]string{"Contact": fmt.Sprintf(`<sip:%s;transport=ws>`, softphone.fakeDomain)}
			responseMsg := inviteMessage.Response(*softphone, 180, dict, "")
			softphone.response(responseMsg)

			softphone.OnInvite(inviteMessage)
		}
	})
}

func (softphone *Softphone) CloseToInvite() {
	softphone.removeMessageListener(softphone.inviteKey)
}
