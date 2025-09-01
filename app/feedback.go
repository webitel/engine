package app

import (
	"context"
	"encoding/base64"
	"github.com/webitel/engine/model"
)

func (app *App) GenerateFeedback(domainId int64, f *model.FeedbackKey) (string, model.AppError) {
	f.DomainId = domainId
	v, err := app.EncryptBytes(f.ToJson())

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(v), nil
}

func (app *App) CreateFeedback(ctx context.Context, key string, f model.Feedback) (model.Feedback, model.AppError) {
	hk, err := app.feedbackHashKey(key)
	if err != nil {
		app.Log.Error(err.Error())
		return model.Feedback{}, model.NewBadRequestError("feedback", "bad request")
	}

	f, err = app.Store.Feedback().Create(ctx, hk, f.Rating, f.Description)
	if err != nil {
		app.Log.Error(err.Error())
		return model.Feedback{}, model.NewBadRequestError("feedback", "bad request")
	}

	return f, nil
}

func (app *App) GetFeedback(ctx context.Context, key string) (model.Feedback, model.AppError) {
	hk, err := app.feedbackHashKey(key)
	if err != nil {
		app.Log.Error(err.Error())
		return model.Feedback{}, model.NewBadRequestError("feedback", "bad request")
	}

	f, err := app.Store.Feedback().Get(ctx, hk)
	if err != nil {
		app.Log.Error(err.Error())
		return model.Feedback{}, model.NewCustomCodeError("feedback", "", err.GetStatusCode())
	}

	return f, nil
}

func (app *App) feedbackHashKey(key string) (model.FeedbackKey, model.AppError) {
	keyBytes, _ := base64.URLEncoding.DecodeString(key)
	v, err := app.DecryptBytes(keyBytes)
	if err != nil {
		return model.FeedbackKey{}, err
	}

	return model.FeedbackKeyFromJson(v)
}
