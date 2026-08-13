package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/webitel/engine/call_manager"
	"github.com/webitel/webitel-go-kit/infra/health"
)

// freeswitchCheck reports whether this node's FreeSWITCH is usable.
func freeswitchCheck(cm call_manager.CallManager) health.Check {
	return func(context.Context) error {
		cli, appErr := cm.CallClient()
		if appErr != nil {
			return fmt.Errorf("freeswitch client unavailable: %w", appErr)
		}

		if !cli.Ready() {
			return errors.New("freeswitch client not ready")
		}

		return nil
	}
}
