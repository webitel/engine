package grpc_api

import (
	"context"
	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
)

type feedback struct {
	*API
	engine.UnsafeFeedbackServiceServer
}

func NewFeedbackApi(api *API) *feedback {
	return &feedback{API: api}
}

func (api *feedback) GenerateFeedback(ctx context.Context, in *engine.GenerateFeedbackRequest) (*engine.GenerateFeedbackResponse, error) {
	key, err := api.app.GenerateFeedback(in.DomainId, &model.FeedbackKey{
		Source:   in.Source,
		SourceId: in.SourceId,
		Payload:  in.Payload,
	})

	if err != nil {
		return nil, err
	}

	return &engine.GenerateFeedbackResponse{
		Key: key,
	}, nil
}

func (api *feedback) GetFeedback(ctx context.Context, in *engine.GetFeedbackRequest) (*engine.Feedback, error) {

	f, err := api.app.GetFeedback(ctx, in.Key)
	if err != nil {
		return nil, err
	}

	return &engine.Feedback{
		Payload:     f.Payload,
		CreatedAt:   f.CreatedAt,
		Rating:      f.Rating,
		Description: f.Description,
	}, nil
}

func (api *feedback) CreateFeedback(ctx context.Context, in *engine.CreateFeedbackRequest) (*engine.Feedback, error) {

	f, err := api.app.CreateFeedback(ctx, in.Key, model.Feedback{
		Rating:      in.Rating,
		Description: in.Description,
	})

	if err != nil {
		return nil, err
	}

	return &engine.Feedback{
		Payload:     f.Payload,
		CreatedAt:   f.CreatedAt,
		Rating:      f.Rating,
		Description: f.Description,
	}, nil

}
