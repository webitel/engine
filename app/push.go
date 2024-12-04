package app

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/wlog"
	"google.golang.org/api/option"
	"sync"
	"time"
)

var firebaseClient *messaging.Client
var apnClient *ApnClient

func initFirebase(key string) error {
	opt := option.WithCredentialsFile(key)
	a, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}
	firebaseClient, err = a.Messaging(context.Background())
	return err
}

func initApn(conf model.PushConfig) error {
	cert, err := ApnCertificate(&conf)
	if err != nil {
		return err
	}
	apnClient = NewApnClient(cert, map[string]string{
		"Content-Type": "application/json",
		"apns-topic":   conf.ApnTopic,
	})
	return nil
}

func (app *App) SendPush(ctx context.Context, r *model.SendPush) (int, model.AppError) {
	var sendAndroid = 0
	var sendAPN = 0
	wg := sync.WaitGroup{}
	if firebaseClient != nil && len(r.Android) != 0 {
		wg.Add(1)
		go func() {
			sendAndroid = pushFirebase(ctx, r)
			wg.Done()
		}()
	}
	if apnClient != nil && len(r.Apple) != 0 {
		wg.Add(1)
		go func() {
			sendAPN = pushApn(ctx, r)
			wg.Done()
		}()
	}

	wg.Wait()

	return sendAndroid + sendAPN, nil
}

func pushApn(ctx context.Context, r *model.SendPush) int {

	var err model.AppError
	var send = 0
	for _, v := range r.Apple {
		err = apnClient.Push(ctx, v, r)
		if err != nil {
			wlog.Error(err.Error(), wlog.Namespace("context"),
				wlog.String("protocol", "apn"),
				wlog.Err(err),
			)
			continue
		}
		send++
	}

	return send
}

func pushFirebase(ctx context.Context, r *model.SendPush) int {

	t := time.Millisecond * time.Duration(r.Expiration)
	count := 0
	priority := "normal"
	if r.Priority > 5 {
		priority = "high"
	}
	res, err := firebaseClient.SendEachForMulticast(ctx, &messaging.MulticastMessage{
		Tokens: r.Android,
		//Data:         r.Data,
		Notification: nil,
		Android: &messaging.AndroidConfig{
			CollapseKey:           "",
			Priority:              priority,
			TTL:                   &t,
			RestrictedPackageName: "",
			Data:                  r.Data,
			Notification:          nil,
			FCMOptions:            nil,
		},
		Webpush: nil,
		APNS:    nil,
	})

	if err != nil {
		wlog.Error(err.Error(), wlog.Namespace("context"),
			wlog.String("protocol", "firebase"),
			wlog.Err(err),
		)
	} else if res != nil {

		if res.FailureCount > 0 {
			for _, v := range res.Responses {
				if !v.Success {
					wlog.Error(v.Error.Error(), wlog.Namespace("context"),
						wlog.String("protocol", "firebase"),
						wlog.Any("response", v.Error),
					)
				}
			}
		}

		count += res.SuccessCount

	}

	return count
}
