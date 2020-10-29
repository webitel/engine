module github.com/webitel/engine

go 1.15

require (
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/consul/api v1.3.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.8.0
	github.com/nicksnyder/go-i18n v1.10.1
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/webitel/call_center v0.0.0-20201016135911-2d0a1ae5af6e
	github.com/webitel/protos/cc v0.0.0-20201027102345-8712c66e378e // indirect
	github.com/webitel/protos/chat v0.0.0-20201027102345-8712c66e378e
	github.com/webitel/protos/engine v0.0.0-20201027102345-8712c66e378e
	github.com/webitel/wlog v0.0.0-20190823170623-8cc283b29e3e
	go.uber.org/atomic v1.6.0
	go.uber.org/ratelimit v0.1.0
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	google.golang.org/genproto v0.0.0-20201021134325-0d71844de594
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/webitel/protos => github.com/webitel/protos v0.0.0-20201027102345-8712c66e378e
