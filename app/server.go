package app

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/webitel/wlog"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	WebSocketRouter *WebSocketRouter
	// RootRouter is the starting point for all HTTP requests to the server.
	RootRouter *mux.Router

	// Router is the starting point for all web, api4 and ws requests to the server. It differs
	// from RootRouter only if the SiteURL contains a /subpath.
	Router *mux.Router

	Server     *http.Server
	ListenAddr *net.TCPAddr

	didFinishListen chan struct{}
}

type RecoveryLogger struct {
}

type CorsWrapper struct {
	router *mux.Router
}

func (cw *CorsWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//TODO
	if r.Header.Get("Origin") == "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	}

	if r.Method == "OPTIONS" {
		w.Header().Set(
			"Access-Control-Allow-Methods",
			strings.Join([]string{"GET", "POST", "PUT", "DELETE"}, ", "))

		w.Header().Set(
			"Access-Control-Allow-Headers",
			strings.Join([]string{r.Header.Get("Access-Control-Request-Headers"), "Access-Control-Allow-Credentials"}, ","))
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	cw.router.ServeHTTP(w, r)
}

func (rl *RecoveryLogger) Println(i ...interface{}) {
	wlog.Error("Please check the std error output for the stack trace")
	wlog.Error(fmt.Sprint(i))
}

func (a *App) StartServer() error {
	wlog.Info("starting server...")
	var handler http.Handler = &CorsWrapper{a.Srv.RootRouter}

	a.Srv.Server = &http.Server{
		Handler:  handlers.RecoveryHandler(handlers.RecoveryLogger(&RecoveryLogger{}), handlers.PrintRecoveryStack(true))(handler),
		ErrorLog: a.Log.StdLog(wlog.String("source", "httpserver")),
	}

	addr := a.Config().WebSocketSettings.Address
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	a.Srv.ListenAddr = listener.Addr().(*net.TCPAddr)
	wlog.Info(fmt.Sprintf("server is listening on %v", listener.Addr().String()))
	a.Srv.didFinishListen = make(chan struct{})

	go func() {
		var err error

		err = a.Srv.Server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			wlog.Critical(fmt.Sprintf("error starting server, err:%v", err))
			time.Sleep(time.Second)
		}
		close(a.Srv.didFinishListen)
	}()

	return nil
}
