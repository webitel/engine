package main

import (
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

//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.fs.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.cc.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.flow.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.logger.yaml
//go:generate go run github.com/bufbuild/buf/cmd/buf@latest generate --template buf/buf.gen.yaml
//go:generate go mod tidy

func main() {
	interruptChan := make(chan os.Signal, 1)
	a, err := app.New()
	if err != nil {
		wlog.Critical("failed to start", wlog.String("error", err.Error()))
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

	var dbg *http.Server

	if a.Config().Dev {
		dbg = setDebug()
	}

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan
	if dbg != nil {
		dbg.Close()
	}
}

func setDebug() *http.Server {
	//debug.SetGCPercent(-1)
	server := &http.Server{Addr: ":8091", Handler: nil}

	go func(s *http.Server) {
		wlog.Info("start debug server on http://localhost:8091/debug/pprof/")
		s.ListenAndServe()
		wlog.Info("stop debug server on http://localhost:8091/debug/pprof/")
	}(server)

	return server
}
