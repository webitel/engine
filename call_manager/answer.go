package call_manager

import (
	"fmt"
)

func (softphone *Softphone) Answer(inviteMessage SipMessage, sdp string) {

	dict := map[string]string{
		"Contact":      fmt.Sprintf("<sip:%s;transport=ws>", softphone.fakeEmail),
		"Content-Type": "application/sdp",
	}
	responseMsg := inviteMessage.Response(*softphone, 200, dict, sdp)
	softphone.response(responseMsg)
}
