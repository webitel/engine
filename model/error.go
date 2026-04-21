package model

// TODO

import (
	"encoding/json"

	"github.com/webitel/engine/pkg/werror"
)

type AppError werror.AppError

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewInternalError(id, details string) AppError {
	return werror.NewInternalError(id, details)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewNotFoundError(id, details string) AppError {
	return werror.NewNotFoundError(id, details)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewBadRequestError(id, details string) AppError {
	return werror.NewBadRequestError(id, details)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewForbiddenError(id, details string) AppError {
	return werror.NewForbiddenError(id, details)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewUnauthorizedError(id, details string) AppError {
	return werror.NewUnauthorizedError(id, details)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
// * NewAutomaticError accepts an code determines in the runtime the status code
func NewCustomCodeError(id, details string, code int) AppError {
	return werror.NewCustomCodeError(id, details, code)
}

func AppErrorFromJson(js string) AppError {
	var err werror.ApplicationError
	_ = json.Unmarshal([]byte(js), &err)
	if err.Id == "" {
		return NewInternalError("model.app_error.json_parse.app_error", js)
	}

	return &err
}
