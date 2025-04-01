package app

import (
	"encoding/json"
	"fmt"
	"github.com/BoRuDar/configuration/v5"
	"github.com/webitel/engine/model"
	"io"
	"os"
	"reflect"
)

const (
	DefaultProviderName = `DefaultProvider`
	DefaultProviderTag  = `default`
)

func (app *App) Config() *model.Config {
	return app.config
}

func (app *App) MaxSocketInboundMsgSize() int {
	return int(app.config.WebSocketSettings.MaxInboundMessageSize)
}

func loadConfig() (*model.Config, error) {
	config, err := configuration.New[model.Config](
		configuration.NewEnvProvider(),
		configuration.NewFlagProvider(),
		NewDefaultProvider(),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to init config: %w", err)
	}

	if config.ConfigFile != nil && *config.ConfigFile != "" {
		var body []byte
		f, err := os.OpenFile(*config.ConfigFile, os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if body, err = io.ReadAll(f); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(body, &config); err != nil {
			return nil, err
		}
	}

	if !config.Log.Console && !config.Log.Otel && len(config.Log.File) == 0 {
		config.Log.Console = true
	}

	// CaseTriggersSettings  : trying to use default AMQP url if config option is empty
	if config.CaseTriggersSettings.BrokerUrl == "" {
		config.CaseTriggersSettings.BrokerUrl = config.MessageQueueSettings.Url
	}

	return config, nil
}

// NewDefaultProvider creates new provider which sets values from `default` tag
// nolint:revive
func NewDefaultProvider() defaultProvider {
	return defaultProvider{}
}

type defaultProvider struct{}

func (defaultProvider) Name() string {
	return DefaultProviderName
}

func (defaultProvider) Tag() string {
	return DefaultProviderTag
}

func (defaultProvider) Init(_ any) error {
	return nil
}

func (dp defaultProvider) Provide(field reflect.StructField, v reflect.Value) error {
	valStr := field.Tag.Get(DefaultProviderTag)
	// allow empty string

	return configuration.SetField(field, v, valStr)
}
