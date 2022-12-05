package app

import (
	"flag"
	"fmt"
	"github.com/webitel/engine/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	appId                   = flag.String("id", "1", "Service id")
	translationsDirectory   = flag.String("translations_directory", "i18n", "Translations directory")
	consulHost              = flag.String("consul", "172.0.0.1:8500", "Host to consul")
	websocketHost           = flag.String("websocket", ":80", "WebSocket server address")
	dataSource              = flag.String("data_source", "postgres://opensips:webitel@postgres:5432/webitel?fallback_application_name=engine&sslmode=disable&connect_timeout=10&search_path=call_center", "Data source")
	amqpSource              = flag.String("amqp", "amqp://webitel:webitel@rabbit:5672?heartbeat=10", "AMQP connection")
	grpcServerPort          = flag.Int("grpc_port", 0, "GRPC port")
	grpcServerAddr          = flag.String("grpc_addr", "", "GRPC host")
	openSipAddr             = flag.String("open_sip_addr", "opensips", "OpenSip address")
	sipPublicProxyAddr      = flag.String("sip_proxy_addr", "", "Public sip proxy address")
	wsSipAddr               = flag.String("ws_sip_addr", "", "Sip websocket address")
	dev                     = flag.Int("dev", 0, "enable dev mode")
	cloudflare              = flag.Int("cloudflare", 0, "use cloudflare")
	minMaskLen              = flag.Int("min_mask_number_len", 0, "Minimum mask length number")
	prefCntMaskLen          = flag.Int("prefix_number_mask_len", 5, "Prefix mask length number")
	suffCntMaskLen          = flag.Int("suffix_number_mask_len", 3, "Suffix mask length number")
	sqlQueryTimeout         = flag.Int("sql_query_timeout", 10, "Sql query timeout sec")
	pingClientInterval      = flag.Int("ping_client_interval", 0, "Interval websocket ping")
	socketMaxInboundMsgSize = flag.String("socket_max_in_msg_size", "256KB", "Maximum inbound size websocket message message")
)

func (app *App) Config() *model.Config {
	return app.config
}

func (app *App) MaxSocketInboundMsgSize() int {
	return app.config.WebSocketSettings.MaxInboundMessageSize
}

func loadConfig() (*model.Config, error) {
	flag.Parse()

	wsMsgInSize, err := model.FromHumanSize(*socketMaxInboundMsgSize)
	if err != nil {
		return nil, err
	}

	config := &model.Config{
		Dev:                   *dev == 1,
		Cloudflare:            *cloudflare == 1,
		PingClientInterval:    *pingClientInterval,
		TranslationsDirectory: *translationsDirectory,
		NodeName:              fmt.Sprintf("engine-%s", *appId),
		MinimumNumberMaskLen:  *minMaskLen,
		PrefixNumberMaskLen:   *prefCntMaskLen,
		SuffixNumberMaskLen:   *suffCntMaskLen,
		DiscoverySettings: model.DiscoverySettings{
			Url: *consulHost,
		},
		WebSocketSettings: model.WebSocketSettings{
			Address:               *websocketHost,
			MaxInboundMessageSize: int(wsMsgInSize),
		},
		ServerSettings: model.ServerSettings{
			Address: *grpcServerAddr,
			Port:    *grpcServerPort,
			Network: "tcp",
		},
		SipSettings: model.SipSettings{
			ServerAddr:  *wsSipAddr,
			Proxy:       *openSipAddr,
			PublicProxy: *sipPublicProxyAddr,
		},
		SqlSettings: model.SqlSettings{
			DriverName:                  model.NewString("postgres"),
			DataSource:                  dataSource,
			MaxIdleConns:                model.NewInt(5),
			MaxOpenConns:                model.NewInt(5),
			ConnMaxLifetimeMilliseconds: model.NewInt(300000),
			QueryTimeout:                sqlQueryTimeout,
			Trace:                       false,
		},
		MessageQueueSettings: model.MessageQueueSettings{
			Url: *amqpSource,
		},
		EmailOAuth: make(map[string]oauth2.Config),
	}

	config.EmailOAuth[model.MailGoogle] = oauth2.Config{
		RedirectURL:  "https://dev.webitel.com/endpoint/oauth2/google/callback",
		ClientID:     "1003527838078-eq1o4od8bvrvquk6a5m8gfkauhria0dj.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint:     google.Endpoint,
	}

	//config.EmailOAuth[model.MailGoogle] = oauth2.Config{
	//	RedirectURL:  "http://dev.webitel.com/endpoint/oauth2/google/callback",
	//	ClientID:     "1003527838078-68qhem81etudj2qol550ruc6qlul5ut3.apps.googleusercontent.com",
	//	ClientSecret: "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",
	//	Scopes:       []string{"https://mail.google.com/"},
	//	Endpoint:     google.Endpoint,
	//}

	return config, nil
}
