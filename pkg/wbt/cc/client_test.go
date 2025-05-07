package client

import (
	"testing"
)

func Test(t *testing.T) {
	t.Log("CC")

	cc := NewCCManager("10.9.8.111:8500")
	cc.Start()
	cc.Agent().Pause(1, 12, "", 14)
	defer cc.Stop()
}
