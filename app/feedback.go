package app

import (
	"context"
	"github.com/webitel/engine/model"
)

func (app *App) GenerateFeedback(domainId int64, f *model.FeedbackKey) (string, model.AppError) {
	f.DomainId = domainId
	v, err := app.EncryptBytes(f.ToJson())

	if err != nil {
		return "", err
	}

	return string(v), nil
}

func (app *App) CreateFeedback(ctx context.Context, key string, f model.Feedback) (model.Feedback, model.AppError) {
	hk, err := app.feedbackHashKey(key)
	if err != nil {
		return model.Feedback{}, err
	}

	f, err = app.Store.Feedback().Create(ctx, hk, f.Rating, f.Description)
	if err != nil {
		return model.Feedback{}, err
	}

	return f, nil
}

func (app *App) GetFeedback(ctx context.Context, key string) (model.Feedback, model.AppError) {
	hk, err := app.feedbackHashKey(key)
	if err != nil {
		return model.Feedback{}, err
	}

	f, err := app.Store.Feedback().Get(ctx, hk)
	if err != nil {
		return model.Feedback{}, err
	}

	return f, nil
}

func (app *App) feedbackHashKey(key string) (model.FeedbackKey, model.AppError) {
	v, err := app.DecryptBytes([]byte(key))
	if err != nil {
		return model.FeedbackKey{}, err
	}

	return model.FeedbackKeyFromJson(v)
}
