//go:build integration

package cc_test

import (
	"testing"

	"github.com/webitel/engine/app/cc"
)

func Test(t *testing.T) {
	t.Log("CC")

	cc := cc.NewCCManager("10.9.8.111:8500")
	cc.Start()
	cc.Agent().Pause(1, 12, "", "", 14)
	defer cc.Stop()
}
