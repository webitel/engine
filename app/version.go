package app

import (
	"fmt"
	"github.com/webitel/engine/model"
)

func (app *App) Version() string {
	return Version()
}

func Version() string {
	return fmt.Sprintf("%s [build:%s]", model.CurrentVersion, model.BuildNumber)
}
