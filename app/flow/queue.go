package flow

import (
	"context"
	"github.com/webitel/engine/gen/workflow"
	"github.com/webitel/engine/pkg/wbt"
)

type QueueApi interface {
	DoDistributeAttempt(in *workflow.DistributeAttemptRequest) (*workflow.DistributeAttemptResponse, error)
	ResultAttempt(in *workflow.ResultAttemptRequest) (*workflow.ResultAttemptResponse, error)
	StartFlow(in *workflow.StartFlowRequest) (string, error)
	StartSyncFlow(in *workflow.StartSyncFlowRequest) (string, error)
	NewProcessing(ctx context.Context, domainId int64, schemaId int, vars map[string]string) (*QueueProcessing, error)
}

type queueApi struct {
	*wbt.Client[workflow.FlowServiceClient]
	processing *wbt.Client[workflow.FlowProcessingServiceClient]
}

func NewQueueApi(api *wbt.Client[workflow.FlowServiceClient], p *wbt.Client[workflow.FlowProcessingServiceClient]) QueueApi {
	return &queueApi{
		Client:     api,
		processing: p,
	}
}

func (q *queueApi) DoDistributeAttempt(in *workflow.DistributeAttemptRequest) (*workflow.DistributeAttemptResponse, error) {
	return q.Api.DistributeAttempt(context.Background(), in)
}

func (q *queueApi) ResultAttempt(in *workflow.ResultAttemptRequest) (*workflow.ResultAttemptResponse, error) {
	return q.Api.ResultAttempt(context.Background(), in)
}

func (q *queueApi) StartFlow(in *workflow.StartFlowRequest) (string, error) {
	res, err := q.Api.StartFlow(context.Background(), in)
	if err != nil {

		return "", err
	}

	return res.Id, nil
}

func (q *queueApi) StartSyncFlow(in *workflow.StartSyncFlowRequest) (string, error) {
	res, err := q.Api.StartSyncFlow(context.Background(), in)
	if err != nil {

		return "", err
	}

	return res.Id, nil
}
