package call_manager

import (
	"github.com/webitel/engine/model"
	"net/http"
)

var (
	NotFoundCall = model.NewAppError("Call", "call.not_found", nil, "call not found",
		http.StatusNotFound)
)
