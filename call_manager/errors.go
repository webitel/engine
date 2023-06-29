package call_manager

import (
	"github.com/webitel/engine/model"
)

var (
	NotFoundCall = model.NewNotFoundError("call.not_found", "call not found")
)
