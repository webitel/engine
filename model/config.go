package model

import (
	"reflect"
)

const (
	DEFAULT_LOCALE = "en"

	DATABASE_DRIVER_POSTGRES = "postgres"
)

type LocalizationSettings struct {
	DefaultServerLocale *string `json:"default_server_locale" default:"en"`
	DefaultClientLocale *string `json:"default_client_locale" default:"en"`
	AvailableLocales    *string `json:"available_locales" default:"en"`
}

type SipSettings struct {
	Proxy       string `json:"proxy" flag:"open_sip_addr|opensips|OpenSIP address" env:"OPEN_SIP_ADDR"`
	ServerAddr  string `json:"server_addr" flag:"ws_sip_addr||Sip websocket address" default:""  env:"WS_SIP_ADDR"`
	PublicProxy string `json:"public_proxy" flag:"sip_proxy_addr||Public sip proxy address" default:"" env:"SIP_PROXY_ADDR"`
}

type LogSettings struct {
	Lvl     string `json:"lvl" flag:"log_lvl|debug|Log level" env:"LOG_LVL"`
	Json    bool   `json:"json" flag:"log_json|false|Log format JSON" env:"LOG_JSON"`
	Otel    bool   `json:"otel" flag:"log_otel|false|Log OTEL" env:"LOG_OTEL"`
	File    string `json:"file" flag:"log_file||Log file directory"  default:"" env:"LOG_FILE"`
	Console bool   `json:"console" flag:"log_console|false|Log console" env:"LOG_CONSOLE"`
}

type Config struct {
	ConfigFile            *string              `json:"-" flag:"config_file||JSON file configuration" default:"" env:"CONFIG_FILE"`
	PresignedCert         string               `json:"presigned_cert" flag:"presigned_cert|/opt/storage/key.pem|Location to pre signed certificate" env:"PRESIGNED_CERT"`
	TranslationsDirectory string               `json:"translations_directory" flag:"translations_directory|i18n|Translations directory" env:"TRANSLATION_DIRECTORY"`
	NodeName              string               `flag:"id|1|Service id" json:"id" env:"ID"`
	DiscoverySettings     DiscoverySettings    `json:"discovery_settings"`
	LocalizationSettings  LocalizationSettings `json:"localization_settings"`
	MessageQueueSettings  MessageQueueSettings `json:"message_queue_settings"`
	SqlSettings           SqlSettings          `json:"sql_settings"`
	AuthCacheExpire       int64                `json:"auth_cache_expire" flag:"auth_cache_expire|30|Auth cache expire in seconds" env:"AUTH_CACHE_EXPIRE"`
	ServerSettings        ServerSettings       `json:"server_settings"`
	WebSocketSettings     WebSocketSettings    `json:"web_socket_settings"`
	Dev                   bool                 `json:"dev" flag:"dev|false|Dev mode" env:"DEV"`
	SipSettings           SipSettings          `json:"sip_settings"`
	Cloudflare            bool                 `json:"cloudflare" flag:"cloudflare|0|Use cloudflare"`
	PingClientInterval    int                  `json:"ping_client_interval" flag:"ping_client_interval|0|Interval websocket ping" env:"PING_CLIENT_INTERVAL"`
	PingClientLatency     bool                 `json:"ping_client_latency" flag:"ping_client_latency|0|Websocket ping latency" env:"PING_CLIENT_LATENCY"`
	MinimumNumberMaskLen  int                  `json:"minimum_number_mask_len" flag:"min_mask_number_len|0|Minimum mask length number" env:"MIN_NUMBER_MASK_LEN"`
	PrefixNumberMaskLen   int                  `json:"prefix_number_mask_len" flag:"prefix_number_mask_len|5|Prefix mask length number" env:"PREFIX_NUMBER_MASK_LEN"`
	SuffixNumberMaskLen   int                  `json:"suffix_number_mask_len" flag:"suffix_number_mask_len|5|Suffix mask length number" env:"SUFFIX_NUMBER_MASK_LEN"`
	//EmailOAuth              map[string]oauth2.Config `json:"email_oauth2,omitempty"`
	MaxMemberCommunications int     `json:"max_member_communications" flag:"max_member_communications|20|Maximum member communications" env:"MAX_MEMBER_COMMUNICATIONS"`
	PublicHostName          *string `json:"public_host" flag:"public_host||Public hostname" default:"" env:"PUBLIC_HOST"`
	Push                    PushConfig
	Log                     LogSettings      `json:"log"`
	TriggersSettings        TriggersSettings `json:"triggers_settings"`
	RTCConfiguration        string           `json:"rtc_configuration" flag:"rtc_configuration||RTCConfiguration" default:"" env:"RTC_CONFIGURATION"`
}

type PushConfig struct {
	FirebaseServiceAccount string `json:"push_firebase" flag:"push_firebase||Firebase service account file location" default:"" env:"PUSH_FIREBASE"`

	ApnHost     string `json:"push_apn_host" flag:"push_apn_host||APN http host"  default:"" env:"PUSH_APN_HOST"`
	ApnCertFile string `json:"push_apn_cert_file" flag:"push_apn_cert_file||APN certificate file location"  default:"" env:"PUSH_APN_CERT_FILE"`
	ApnKeyFile  string `json:"push_apn_key_file" flag:"push_apn_key_file||APN key file location"  default:"" env:"PUSH_APN_KEY_FILE"`
	ApnTopic    string `json:"push_apn_topic" flag:"push_apn_topic|com.webitel.webitel-ios.voip|APN topic" env:"PUSH_APN_TOPIC"`
}

type DiscoverySettings struct {
	Url string `json:"url" flag:"consul|172.0.0.1:8500|Host to consul" env:"CONSUL"`
}

type MessageQueueSettings struct {
	Url string `json:"url" flag:"amqp|amqp://admin:admin@rabbit:5672?heartbeat=10|AMQP connection" env:"AMQP"`
}

type ServerSettings struct {
	Address        string   `json:"address" flag:"grpc_addr||GRPC host" default:"" env:"GRPC_ADDR"`
	Port           int      `json:"port" flag:"grpc_port|0|GRPC port" env:"GRPC_PORT"`
	Network        string   `json:"network" flag:"grpc_network|tcp|GRPC network" env:"GRPC_NETWORK"`
	MaxMessageSize ByteSize `json:"max_message_size" flag:"grpc_max_message_size|16MB|Maximum GRPC message size" env:"GRPC_MAX_MESSAGE_SIZE"`
}

type ByteSize int

type WebSocketSettings struct {
	Address               string   `json:"address" flag:"websocket|:80|WebSocket server address" env:"WEBSOCKET"`
	MaxInboundMessageSize ByteSize `json:"max_inbound_message_size" flag:"socket_max_in_msg_size|256KB|Maximum inbound size websocket message message" env:"SOCKET_MAX_IN_MSG_SIZE"`
}

type SqlSettings struct {
	DriverName                  *string `json:"driver_name" flag:"sql_driver_name|postgres|" env:"SQL_DRIVER_NAME"`
	DataSource                  *string `json:"data_source" flag:"data_source|postgres://postgres:postgres@postgres:5432/webitel?fallback_application_name=engine&sslmode=disable&connect_timeout=10&search_path=call_center|Data source" env:"DATA_SOURCE"`
	DataSourceReplicas          string  `json:"data_source_replicas" flag:"sql_data_source_replicas||sql replicas" default:"" env:"SQL_DATA_SOURCE_REPLICAS"`
	MaxIdleConns                *int    `json:"max_idle_conns" flag:"sql_max_idle_conns|5|Maximum idle connections" env:"SQL_MAX_IDLE_CONNS"`
	MaxOpenConns                *int    `json:"max_open_conns" flag:"sql_max_open_conns|5|Maximum open connections" env:"SQL_MAX_OPEN_CONNS"`
	ConnMaxLifetimeMilliseconds *int    `json:"conn_max_lifetime_milliseconds" flag:"sql_conn_max_lifetime_milliseconds|300000|Connection maximum lifetime milliseconds" env:"SQL_LIFETIME_MILLISECONDS"`
	Trace                       bool    `json:"trace" flag:"sql_trace|false|Trace SQL" env:"SQL_TRACE"`
	Log                         bool    `json:"log" flag:"sql_log|false|Log SQL" env:"SQL_LOG"`
	QueryTimeout                *int    `json:"query_timeout" flag:"sql_query_timeout|10|Sql query timeout seconds" env:"QUERY_TIMEOUT"`
}

type TriggersSettings struct {
	Enabled   bool   `json:"enabled" flag:"trigger_enabled|true|Enable trigger" env:"TRIGGER_ENABLED"`
	BrokerUrl string `json:"broker_url" flag:"broker_url||Broker for CaseTriggers"  default:"" env:"TRIGGER_BROKER_URL"`
	Exchange  string `json:"exchange" flag:"triggers_exchange|cases|Exchange name for triggers cases" env:"TRIGGERS_EXCHANGE"`
	Queue     string `json:"queue" flag:"triggers_queue|engine.trigger|Queue name for triggers" env:"TRIGGERS_QUEUE"`
}

func (c *Config) IsValid() AppError {
	if len(c.WebSocketSettings.Address) < 3 {
		return NewBadRequestError("model.config.is_valid.websocket_address.app_error", "")
	}

	if len(c.DiscoverySettings.Url) < 3 {
		return NewBadRequestError("model.config.is_valid.discovery_url.app_error", "")
	}

	if c.SqlSettings.DataSource == nil || len(*c.SqlSettings.DataSource) < 3 {
		return NewBadRequestError("model.config.is_valid.sql_datasource.app_error", "")
	}
	return nil
}

func (s *ByteSize) SetField(_ reflect.StructField, val reflect.Value, valStr string) error {
	v, err := FromHumanSize(valStr)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(ByteSize(v)))
	return nil
}

func (s *LocalizationSettings) SetDefaults() {
	if s.DefaultServerLocale == nil {
		s.DefaultServerLocale = NewString(DEFAULT_LOCALE)
	}

	if s.DefaultClientLocale == nil {
		s.DefaultClientLocale = NewString(DEFAULT_LOCALE)
	}

	if s.AvailableLocales == nil {
		s.AvailableLocales = NewString("")
	}
}

func (c *PushConfig) ValidApn() bool {
	return !(len(c.ApnCertFile) == 0 || len(c.ApnKeyFile) == 0 || len(c.ApnTopic) == 0)
}
