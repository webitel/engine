package model

import (
	"encoding/json"
	"fmt"
	"net/http"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
)

var translateFunc goi18n.TranslateFunc = nil

func AppErrorInit(t goi18n.TranslateFunc) {
	translateFunc = t
}

type AppError interface {
	// * SetTranslationParams represents the parameters that will be passed to the translation function
	SetTranslationParams(map[string]any) AppError
	GetTranslationParams() map[string]any
	// * SetAppearedIn represents the optional parameter that can be used in the service
	SetAppearedIn(where string) AppError
	GetAppearedIn() string
	// * SetStatusCode represents the status code of error
	SetStatusCode(int) AppError
	GetStatusCode() int
	SetDetailedError(string)
	GetDetailedError() string
	SetRequestId(string)
	GetRequestId() string
	GetId() string

	Error() string
	Translate(T goi18n.TranslateFunc)
	SystemMessage(goi18n.TranslateFunc) string
	ToJson() string
	String() string
}

type ApplicationError struct {
	Id            string `json:"id"`
	Where         string `json:"where"`
	Message       string `json:"status"`               // Message to be display to the end user without debugging information
	DetailedError string `json:"detail"`               // Internal error string to help the developer
	RequestId     string `json:"request_id,omitempty"` // The RequestId that's also set in the header
	Status        int    `json:"code,omitempty"`       // The http status code
	params        map[string]interface{}
}

func (er *ApplicationError) SetTranslationParams(params map[string]any) AppError {
	er.params = params
	return er
}
func (er *ApplicationError) GetTranslationParams() map[string]any {
	return er.params
}

func (er *ApplicationError) SetAppearedIn(where string) AppError {
	er.Where = where
	return er
}

func (er *ApplicationError) GetAppearedIn() string {
	return er.Where
}

func (er *ApplicationError) SetStatusCode(code int) AppError {
	er.Status = code
	return er
}

func (er *ApplicationError) GetStatusCode() int {
	return er.Status
}

func (er *ApplicationError) Error() string {
	var where string
	if er.Where != "" {
		where = er.Where + ": "
	}
	return fmt.Sprintf("%s%s, %s", where, er.Message, er.DetailedError)
}
func (er *ApplicationError) SetDetailedError(details string) {
	er.DetailedError = details
}

func (er *ApplicationError) GetDetailedError() string {
	return er.DetailedError
}

func (er *ApplicationError) Translate(T goi18n.TranslateFunc) {
	if T == nil {
		er.Message = er.Id
		return
	}

	if er.params == nil {
		er.Message = T(er.Id)
	} else {
		er.Message = T(er.Id, er.params)
	}
}

func (er *ApplicationError) SystemMessage(T goi18n.TranslateFunc) string {
	if er.params == nil {
		return T(er.Id)
	} else {
		return T(er.Id, er.params)
	}
}

func (er *ApplicationError) SetRequestId(id string) {
	er.RequestId = id
}

func (er *ApplicationError) GetRequestId() string {
	return er.RequestId
}

func (er *ApplicationError) GetId() string {
	return er.Id
}

func (er *ApplicationError) ToJson() string {
	b, _ := json.Marshal(er)
	return string(b)
}

func (er *ApplicationError) String() string {
	if er.Id == er.Message && er.DetailedError != "" {
		return er.DetailedError
	}

	return er.Message
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewInternalError(id string, details string) AppError {
	return newAppError(id, details).SetStatusCode(http.StatusInternalServerError)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewNotFoundError(id string, details string) AppError {
	return newAppError(id, details).SetStatusCode(http.StatusNotFound)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewBadRequestError(id string, details string) AppError {
	return newAppError(id, details).SetStatusCode(http.StatusBadRequest)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewForbiddenError(id string, details string) AppError {
	return newAppError(id, details).SetStatusCode(http.StatusForbidden)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
func NewUnauthorizedError(id string, details string) AppError {
	return newAppError(id, details).SetStatusCode(http.StatusUnauthorized)
}

// ! Id should be built like this written in the snake case --  *package*.*file*.*function*.*in what stage of function error occured*.*what happened*
// * NewAutomaticError accepts an code determines in the runtime the status code
func NewCustomCodeError(id string, details string, code int) AppError {
	if code > 511 || code < 100 {
		code = http.StatusInternalServerError
	}
	return newAppError(id, details).SetStatusCode(code)
}

func newAppError(id string, details string) AppError {
	return &ApplicationError{Id: id, Message: id, DetailedError: details}
}
