package model

import (
	"reflect"

	"golang.org/x/oauth2"
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

type B2BSettings struct {
	Addr string `json:"b2b_addr" flag:"b2b_addr||B2B listen address"`
}

type SipSettings struct {
	Proxy       string `json:"proxy" flag:"open_sip_addr|opensips|OpenSIP address"`
	ServerAddr  string `json:"server_addr" flag:"ws_sip_addr||Sip websocket address"`
	PublicProxy string `json:"public_proxy" flag:"sip_proxy_addr||Public sip proxy address"`
}

type LogSettings struct {
	Lvl  string `json:"lvl" flag:"log_lvl|debug|Log level"`
	Json bool   `json:"json" flag:"log_json|1|Log format JSON"`
	File string `json:"file" flag:"log_file|/var/log/webitel/engine|Log file directory"`
}

type Config struct {
	ConfigFile              *string                  `json:"-" flag:"config_file||JSON file configuration"`
	PresignedCert           string                   `json:"presigned_cert" flag:"presigned_cert|/opt/storage/key.pem|Location to pre signed certificate"`
	TranslationsDirectory   string                   `json:"translations_directory" flag:"translations_directory|i18n|Translations directory"`
	NodeName                string                   `flag:"id|1|Service id" json:"id"`
	DiscoverySettings       DiscoverySettings        `json:"discovery_settings"`
	LocalizationSettings    LocalizationSettings     `json:"localization_settings"`
	MessageQueueSettings    MessageQueueSettings     `json:"message_queue_settings"`
	SqlSettings             SqlSettings              `json:"sql_settings"`
	AuthCacheExpire         int64                    `json:"auth_cache_expire" flag:"auth_cache_expire|30|Auth cache expire in seconds"`
	ServerSettings          ServerSettings           `json:"server_settings"`
	WebSocketSettings       WebSocketSettings        `json:"web_socket_settings"`
	Dev                     bool                     `json:"dev" flag:"dev|false|Dev mode"`
	SipSettings             SipSettings              `json:"sip_settings"`
	Cloudflare              bool                     `json:"cloudflare" flag:"cloudflare|0|Use cloudflare"`
	PingClientInterval      int                      `json:"ping_client_interval" flag:"ping_client_interval|0|Interval websocket ping"`
	PingClientLatency       bool                     `json:"ping_client_latency" flag:"ping_client_latency|0|Websocket ping latency"`
	MinimumNumberMaskLen    int                      `json:"minimum_number_mask_len" flag:"min_mask_number_len|0|Minimum mask length number"`
	PrefixNumberMaskLen     int                      `json:"prefix_number_mask_len" flag:"prefix_number_mask_len|5|Prefix mask length number"`
	SuffixNumberMaskLen     int                      `json:"suffix_number_mask_len" flag:"suffix_number_mask_len|5|Suffix mask length number"`
	EmailOAuth              map[string]oauth2.Config `json:"email_oauth2,omitempty"`
	MaxMemberCommunications int                      `json:"max_member_communications" flag:"max_member_communications|20|Maximum member communications"`
	PublicHostName          *string                  `json:"public_host" flag:"public_host||Public hostname"`
	B2BSettings             B2BSettings
	Push                    PushConfig
	Log                     LogSettings `json:"log"`
}

type PushConfig struct {
	FirebaseServiceAccount string `json:"push_firebase" flag:"push_firebase||Firebase service account file location"`

	ApnCertFile string `json:"push_apn_cert_file" flag:"push_apn_cert_file||APN certificate file location"`
	ApnKeyFile  string `json:"push_apn_key_file" flag:"push_apn_key_file||APN key file location"`
	ApnTopic    string `json:"push_apn_topic" flag:"push_apn_topic|com.webitel.webitel-ios.voip|APN topic"`
}

type DiscoverySettings struct {
	Url string `json:"url" flag:"consul|172.0.0.1:8500|Host to consul"`
}

type MessageQueueSettings struct {
	Url string `json:"url" flag:"amqp|amqp://webitel:webitel@rabbit:5672?heartbeat=10|AMQP connection"`
}

type ServerSettings struct {
	Address        string   `json:"address" flag:"grpc_addr||GRPC host"`
	Port           int      `json:"port" flag:"grpc_port|0|GRPC port"`
	Network        string   `json:"network" flag:"grpc_network|tcp|GRPC network"`
	MaxMessageSize ByteSize `json:"max_message_size" flag:"grpc_max_message_size|16MB|Maximum GRPC message size"`
}

type ByteSize int

type WebSocketSettings struct {
	Address               string   `json:"address" flag:"websocket|:80|WebSocket server address"`
	MaxInboundMessageSize ByteSize `json:"max_inbound_message_size" flag:"socket_max_in_msg_size|256KB|Maximum inbound size websocket message message"`
}

type SqlSettings struct {
	DriverName                  *string  `json:"driver_name" flag:"sql_driver_name|postgres|"`
	DataSource                  *string  `json:"data_source" flag:"data_source|postgres://opensips:webitel@postgres:5432/webitel?fallback_application_name=engine&sslmode=disable&connect_timeout=10&search_path=call_center|Data source"`
	DataSourceReplicas          []string `json:"data_source_replicas" flag:"sql_data_source_replicas" default:""`
	MaxIdleConns                *int     `json:"max_idle_conns" flag:"sql_max_idle_conns|5|Maximum idle connections"`
	MaxOpenConns                *int     `json:"max_open_conns" flag:"sql_max_open_conns|5|Maximum open connections"`
	ConnMaxLifetimeMilliseconds *int     `json:"conn_max_lifetime_milliseconds" flag:"sql_conn_max_lifetime_milliseconds|300000|Connection maximum lifetime milliseconds"`
	Trace                       bool     `json:"trace" flag:"sql_trace|false|Trace SQL"`
	QueryTimeout                *int     `json:"query_timeout" flag:"sql_query_timeout|10|Sql query timeout seconds"`
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
