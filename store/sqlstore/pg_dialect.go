package sqlstore

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"

	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
)

const ForeignKeyViolationErrorCode = pq.ErrorCode("23503")
const DuplicationViolationErrorCode = pq.ErrorCode("23505")
const FromTriggerValidationErrorCode = pq.ErrorCode("09000")

type PostgresJSONDialect struct {
	gorp.PostgresDialect
}

func (d PostgresJSONDialect) ToSqlType(val reflect.Type, maxsize int, isAutoIncr bool) string {
	if val == reflect.TypeOf(model.StringInterface{}) {
		return "JSONB"
	}
	return d.PostgresDialect.ToSqlType(val, maxsize, isAutoIncr)
}

func messageFromErr(err error) string {
	switch e := err.(type) {
	case *pq.Error:
		return e.Detail
	default:
		return e.Error()
	}
}

func isNoRowsError(err error) bool { return errors.Is(err, sql.ErrNoRows) }

func extractCodeFromErr(err error) int {
	code := http.StatusInternalServerError

	if err == sql.ErrNoRows {
		code = http.StatusNotFound
	} else if e, ok := err.(*pq.Error); ok {
		switch e.Code {
		case ForeignKeyViolationErrorCode, DuplicationViolationErrorCode, FromTriggerValidationErrorCode:
			code = http.StatusBadRequest
		}
	}
	return code
}

func isDuplicationViolationErrorCode(err error) bool {
	if e, ok := err.(*pq.Error); ok {
		return e.Code == DuplicationViolationErrorCode
	}

	return false
}
