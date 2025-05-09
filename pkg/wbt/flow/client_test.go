package flow

import (
	"fmt"
	"github.com/webitel/engine/pkg/wbt/gen/workflow"
	"testing"
)

var consulAddr = "10.9.8.111:8500"

func TestFlow(t *testing.T) {
	f := NewFlowManager(consulAddr)
	err := f.Start()
	if err != nil {
		panic(err.Error())
	}
	var job string

	job, err = f.Queue().StartSyncFlow(&workflow.StartSyncFlowRequest{
		SchemaId:   1302,
		DomainId:   1,
		TimeoutSec: 0,
		Variables:  nil,
		Scope:      nil,
	})

	fmt.Println(job)

	if err != nil {
		panic(err.Error())
	}
}
