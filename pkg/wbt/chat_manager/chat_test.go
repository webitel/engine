package chat_manager

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	t.Log("CC")

	cc := NewChatManager("10.9.8.111:8500")
	cc.Start()
	cli, err := cc.Client()
	if err != nil {
		panic(err.Error())
	}

	res, err := cli.ListActive("IHOR", 1, 10, 1, 10)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(res.Items)

	defer cc.Stop()
}
