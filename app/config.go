package app

import (
	"encoding/json"
	"github.com/BoRuDar/configuration/v4"
	"github.com/webitel/engine/model"
	"io"
	"os"
)

func (app *App) Config() *model.Config {
	return app.config
}

func (app *App) MaxSocketInboundMsgSize() int {
	return int(app.config.WebSocketSettings.MaxInboundMessageSize)
}

func loadConfig() (*model.Config, error) {
	var config model.Config
	configurator := configuration.New(
		&config,
		configuration.NewEnvProvider(),
		configuration.NewFlagProvider(),
		configuration.NewDefaultProvider(),
	).SetOptions(configuration.OnFailFnOpt(func(err error) {
		//log.Println(err)
	}))

	if err := configurator.InitValues(); err != nil {
		//return nil, err
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

	return &config, nil
}
