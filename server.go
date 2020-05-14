package main

import (
	"fmt"
	"github.com/webitel/engine/apis"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/grpc_api"
	"github.com/webitel/engine/wsapi"
	"github.com/webitel/wlog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	interruptChan := make(chan os.Signal, 1)
	a, err := app.New()
	wlog.Info(fmt.Sprintf("server build version: %s", app.Version()))
	if err != nil {
		wlog.Critical(err.Error())
		return
	}
	defer a.Shutdown()

	serverErr := a.StartServer()
	if serverErr != nil {
		wlog.Critical(serverErr.Error())
		return
	}

	wsapi.Init(a, a.Srv.WebSocketRouter)
	apis.Init(a, a.Srv.Router)
	grpc_api.Init(a, a.GrpcServer.Server())

	if err := a.StartGrpcServer(); err != nil {
		panic(err.Error())
	}

	if a.Config().Dev {
		setDebug()
	}

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan
}

func setDebug() {
	//debug.SetGCPercent(-1)

	go func() {
		wlog.Info("start debug server on http://localhost:8091/debug/pprof/")
		err := http.ListenAndServe(":8091", nil)
		wlog.Info(err.Error())
	}()
}
