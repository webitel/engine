module github.com/webitel/engine

go 1.15

require (
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/consul/api v1.12.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.5
	github.com/nicksnyder/go-i18n v1.10.1
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/streadway/amqp v1.0.0
	github.com/webitel/call_center v0.0.0-20220503134055-28d54db0bc98
	github.com/webitel/protos/cc v0.0.0-20220428115356-35297e3b1bb4
	github.com/webitel/protos/engine v0.0.0-20220428115356-35297e3b1bb4
	github.com/webitel/wlog v0.0.0-20190823170623-8cc283b29e3e
	go.uber.org/atomic v1.9.0
	go.uber.org/ratelimit v0.2.0
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4
	google.golang.org/grpc v1.46.0
)

require (
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/poy/onpar v1.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
