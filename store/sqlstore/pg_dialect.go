package sqlstore

import (
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/call_center/model"
	"reflect"
)

const ForeignKeyViolationErrorCode = pq.ErrorCode("23503")

type PostgresJSONDialect struct {
	gorp.PostgresDialect
}

func (d PostgresJSONDialect) ToSqlType(val reflect.Type, maxsize int, isAutoIncr bool) string {
	if val == reflect.TypeOf(model.StringInterface{}) {
		return "JSONB"
	}
	return d.PostgresDialect.ToSqlType(val, maxsize, isAutoIncr)
}
