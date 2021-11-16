package model

import "net/http"

const (
	DEFAULT_LOCALE = "en"

	DATABASE_DRIVER_POSTGRES = "postgres"
)

type LocalizationSettings struct {
	DefaultServerLocale *string
	DefaultClientLocale *string
	AvailableLocales    *string
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

type SipSettings struct {
	ServerAddr  string
	PublicProxy string
	Proxy       string
}

type Config struct {
	TranslationsDirectory string               `json:"translations_directory"`
	NodeName              string               `json:"node_name"`
	DiscoverySettings     DiscoverySettings    `json:"discovery_settings"`
	LocalizationSettings  LocalizationSettings `json:"localization_settings"`
	MessageQueueSettings  MessageQueueSettings `json:"message_queue_settings"`
	SqlSettings           SqlSettings          `json:"sql_settings"`
	ServerSettings        ServerSettings       `json:"server_settings"`
	WebSocketSettings     WebSocketSettings    `json:"web_socket_settings"`
	Dev                   bool                 `json:"dev"`
	SipSettings           SipSettings          `json:"sip_settings"`
	Cloudflare            bool                 `json:"cloudflare"`
	MinimumNumberMaskLen  int                  `json:"minimum_number_mask_len"`
}

type DiscoverySettings struct {
	Url string
}

type MessageQueueSettings struct {
	Url string
}

type ServerSettings struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Network string `json:"network"`
}

type WebSocketSettings struct {
	Address string `json:"address"`
}

type SqlSettings struct {
	DriverName                  *string
	DataSource                  *string
	DataSourceReplicas          []string
	DataSourceSearchReplicas    []string
	MaxIdleConns                *int
	ConnMaxLifetimeMilliseconds *int
	MaxOpenConns                *int
	Trace                       bool
	AtRestEncryptKey            string
	QueryTimeout                *int
}

func (c *Config) IsValid() *AppError {
	if len(c.WebSocketSettings.Address) < 3 {
		return NewAppError("Config.IsValid", "model.config.is_valid.websocket_address.app_error", nil, "", http.StatusBadRequest)
	}

	if len(c.DiscoverySettings.Url) < 3 {
		return NewAppError("Config.IsValid", "model.config.is_valid.discovery_url.app_error", nil, "", http.StatusBadRequest)
	}

	if c.SqlSettings.DataSource == nil || len(*c.SqlSettings.DataSource) < 3 {
		return NewAppError("Config.IsValid", "model.config.is_valid.sql_datasource.app_error", nil, "", http.StatusBadRequest)
	}
	return nil
}
