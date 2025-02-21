module github.com/webitel/engine

go 1.23.0

toolchain go1.23.4

require (
	buf.build/gen/go/webitel/cc/protocolbuffers/go v1.36.5-20250220080817-337dbf2ba82b.1
	buf.build/gen/go/webitel/chat/grpc/go v1.5.1-20250219211829-285fd066a972.2
	buf.build/gen/go/webitel/chat/protocolbuffers/go v1.36.5-20250219211829-285fd066a972.1
	buf.build/gen/go/webitel/engine/grpc/go v1.5.1-20250220123754-4f4595344295.2
	buf.build/gen/go/webitel/engine/protocolbuffers/go v1.36.5-20250220123754-4f4595344295.1
	buf.build/gen/go/webitel/fs/grpc/go v1.5.1-20240527133230-b0088245838d.2
	buf.build/gen/go/webitel/fs/protocolbuffers/go v1.36.5-20240527133230-b0088245838d.1
	buf.build/gen/go/webitel/logger/grpc/go v1.5.1-20250128105802-aaacc0377b27.2
	buf.build/gen/go/webitel/logger/protocolbuffers/go v1.36.5-20250128105802-aaacc0377b27.1
	buf.build/gen/go/webitel/webitel-go/grpc/go v1.5.1-20250218105124-2ee3869e4b3a.2
	buf.build/gen/go/webitel/webitel-go/protocolbuffers/go v1.36.5-20250218105124-2ee3869e4b3a.1
	buf.build/gen/go/webitel/workflow/protocolbuffers/go v1.36.5-20240527133229-fa5be56493fd.1
	firebase.google.com/go/v4 v4.15.2
	github.com/BoRuDar/configuration/v4 v4.5.1
	github.com/Masterminds/squirrel v1.5.4
	github.com/XSAM/otelsql v0.37.0
	github.com/ghettovoice/gosip v0.0.0-20250206102957-99cfb663cf0c
	github.com/go-gorp/gorp v2.2.0+incompatible
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/consul/api v1.31.2
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/mbobakov/grpc-consul-resolver v1.5.3
	github.com/nicksnyder/go-i18n v1.10.3
	github.com/pborman/uuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/sirupsen/logrus v1.9.3
	github.com/tevino/abool v1.2.0
	github.com/webitel/call_center v0.0.0-20250220082307-bc120c121b1b
	github.com/webitel/flow_manager v0.0.0-20250221080746-58198c489032
	github.com/webitel/webitel-go-kit v0.0.13-0.20240908192731-3abe573c0e41
	github.com/webitel/wlog v0.0.0-20240909100805-822697e17a45
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/sdk v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
	go.uber.org/atomic v1.11.0
	go.uber.org/ratelimit v0.3.1
	golang.org/x/net v0.35.0
	golang.org/x/oauth2 v0.26.0
	golang.org/x/sync v0.11.0
	google.golang.org/api v0.222.0
	google.golang.org/grpc v1.70.0
)

require (
	buf.build/gen/go/grpc-ecosystem/grpc-gateway/protocolbuffers/go v1.36.5-20241220201140-4c5ba75caaf8.1 // indirect
	buf.build/gen/go/webitel/cc/grpc/go v1.5.1-20250220080817-337dbf2ba82b.2 // indirect
	buf.build/gen/go/webitel/workflow/grpc/go v1.5.1-20240527133229-fa5be56493fd.2 // indirect
	cel.dev/expr v0.21.2 // indirect
	cloud.google.com/go v0.118.3 // indirect
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/firestore v1.18.0 // indirect
	cloud.google.com/go/iam v1.4.0 // indirect
	cloud.google.com/go/longrunning v0.6.4 // indirect
	cloud.google.com/go/monitoring v1.24.0 // indirect
	cloud.google.com/go/storage v1.50.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.26.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.50.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.50.0 // indirect
	github.com/MicahParks/keyfunc v1.9.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/discoviking/fsm v0.0.0-20150126104936-f4a273feecca // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/form v3.1.4+incompatible // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/serf v0.10.2 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/onsi/gomega v1.26.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/poy/onpar v1.1.2 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.9.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.34.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.34.0 // indirect
	go.opentelemetry.io/otel/log v0.10.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.10.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	google.golang.org/appengine/v2 v2.0.6 // indirect
	google.golang.org/genproto v0.0.0-20250219182151-9fdb1cabc7b2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250219182151-9fdb1cabc7b2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250219182151-9fdb1cabc7b2 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
