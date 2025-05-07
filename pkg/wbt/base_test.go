package wbt

import (
	"context"
	"fmt"
	"github.com/webitel/engine/pkg/wbt/gen/fs"
	"testing"
)

var consulAddr = "10.9.8.111:8500"

func TestCli(t *testing.T) {
	c, _ := NewClient(consulAddr, "freeswitch", fs.NewApiClient)
	testFsDirect(c)
	testFsRR(c)
}

func testFsDirect(c *Client[fs.ApiClient]) {
	ctx := c.StaticHost(context.Background(), "dev")
	res, err := c.Api.Execute(ctx, &fs.ExecuteRequest{
		Command: "version",
		Args:    "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
}

func testFsRR(c *Client[fs.ApiClient]) {
	ctx := context.Background()
	res, err := c.Api.Execute(ctx, &fs.ExecuteRequest{
		Command: "version",
		Args:    "",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(res)
}
