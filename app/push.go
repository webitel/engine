package app

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
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
		go pushFirebase(ctx, r, &wg, &sendAndroid)
	}
	if apnClient != nil && len(r.Apple) != 0 {
		wg.Add(1)
		go pushApn(ctx, r, &wg, &sendAPN)
	}

	wg.Wait()

	return sendAndroid + sendAPN, nil
}

func pushApn(ctx context.Context, r *model.SendPush, wg *sync.WaitGroup, send *int) {
	go func() {
		wg.Done()
	}()

	var err model.AppError
	for _, v := range r.Apple {
		err = apnClient.Development().Push(ctx, v, r)
		if err != nil {
			wlog.Error(err.Error())
			continue
		}
		*send++
	}
}

func pushFirebase(ctx context.Context, r *model.SendPush, wg *sync.WaitGroup, send *int) {
	defer func() {
		wg.Done()
	}()

	t := time.Millisecond * time.Duration(r.Expiration)
	priority := "normal"
	if r.Priority > 5 {
		priority = "high"
	}
	res, err := firebaseClient.SendMulticast(ctx, &messaging.MulticastMessage{
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
		wlog.Error(fmt.Sprintf("firebase send error: %v", err.Error()))
	} else {
		*send = res.SuccessCount
	}
}
