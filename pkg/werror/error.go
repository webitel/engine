package werror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AppError interface {
	// * SetTranslationParams represents the parameters that will be passed to the translation function
	SetTranslationParams(map[string]any) AppError
	GetTranslationParams() map[string]any
	// * SetStatusCode represents the status code of error
	SetStatusCode(int) AppError
	GetStatusCode() int
	SetDetailedError(string)
	GetDetailedError() string
	SetRequestId(string)
	GetRequestId() string
	GetId() string

	Error() string
	ToJson() string
	String() string
}

type ApplicationError struct {
	Id            string `json:"id"`
	Where         string `json:"where,omitempty"`
	Status        string `json:"status"`               // Message to be display to the end user without debugging information
	DetailedError string `json:"detail"`               // Internal error string to help the developer
	RequestId     string `json:"request_id,omitempty"` // The RequestId that's also set in the header
	StatusCode    int    `json:"code,omitempty"`       // The http status code
	params        map[string]interface{}
}

func (er *ApplicationError) SetTranslationParams(params map[string]any) AppError {
	er.params = params
	return er
}

func (er *ApplicationError) GetTranslationParams() map[string]any {
	return er.params
}

func (er *ApplicationError) SetStatusCode(code int) AppError {
	er.StatusCode = code
	er.Status = http.StatusText(er.StatusCode)
	return er
}

func (er *ApplicationError) GetStatusCode() int {
	return er.StatusCode
}

func (er *ApplicationError) Error() string {
	var where string
	if er.Where != "" {
		where = er.Where + ": "
	}
	return fmt.Sprintf("%s%s, %s", where, er.Status, er.DetailedError)
}
func (er *ApplicationError) SetDetailedError(details string) {
	er.DetailedError = details
}

func (er *ApplicationError) GetDetailedError() string {
	return er.DetailedError
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
	if er.Id == er.Status && er.DetailedError != "" {
		return er.DetailedError
	}

	return er.Status
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
	return &ApplicationError{Id: id, Status: id, DetailedError: details}
}

func AppErrorFromJson(js string) *ApplicationError {
	var err ApplicationError
	json.Unmarshal([]byte(js), &err)
	if err.Id == "" {
		return nil
	}

	return &err
}
