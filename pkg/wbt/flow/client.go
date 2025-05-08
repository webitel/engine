package flow

import (
	"github.com/webitel/engine/pkg/wbt"
	"github.com/webitel/engine/pkg/wbt/gen/workflow"
	"sync"

	"github.com/webitel/wlog"
)

type FlowManager interface {
	Start() error
	Stop()

	Queue() QueueApi
}

type flowManager struct {
	startOnce  sync.Once
	consulAddr string

	flowCli       *wbt.Client[workflow.FlowServiceClient]
	processingCli *wbt.Client[workflow.FlowProcessingServiceClient]
	queue         QueueApi
}

func NewFlowManager(consulAddr string) FlowManager {
	fm := &flowManager{
		consulAddr: consulAddr,
	}

	return fm
}

func (f *flowManager) Start() error {
	wlog.Debug("starting flow manager service")
	var err error

	f.startOnce.Do(func() {
		f.flowCli, err = wbt.NewClient(f.consulAddr, wbt.FlowServiceName, workflow.NewFlowServiceClient)
		if err != nil {
			return
		}

		f.processingCli, err = wbt.NewClient(f.consulAddr, wbt.FlowServiceName, workflow.NewFlowProcessingServiceClient)
		if err != nil {
			return
		}

		f.queue = NewQueueApi(f.flowCli, f.processingCli)

	})
	return err
}

func (f *flowManager) Stop() {
	if f.flowCli != nil {
		_ = f.flowCli.Close()
	}
	if f.processingCli != nil {
		_ = f.processingCli.Close()
	}
}

func (f *flowManager) Queue() QueueApi {
	return f.queue
}
