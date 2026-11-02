//go:build integration

package flow_test

import (
	"testing"

	"github.com/webitel/engine/app/flow"
	"github.com/webitel/engine/gen/workflow"
)

var consulAddr = "10.9.8.111:8500"

func TestFlow(t *testing.T) {
	f := flow.NewFlowManager(consulAddr)
	err := f.Start()
	if err != nil {
		panic(err.Error())
	}

	_, err = f.Queue().StartSyncFlow(&workflow.StartSyncFlowRequest{
		SchemaId:   1302,
		DomainId:   1,
		TimeoutSec: 0,
		Variables:  nil,
		Scope:      nil,
	})

	if err != nil {
		panic(err.Error())
	}
}
