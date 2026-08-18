package sqlstore

import (
	"bytes"
	"context"
	"database/sql"
	drv "database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/webitel/crypto/cryptostore/schema"
)

// encryptBytes quietly encrypts given [plain] data value.
// Optional ctx.. MAY describe <[schema.]table> <column> [<jpath>...] of the data field.
func encryptBytes(plain []byte, tableColumn ...string) []byte {
	if len(plain) == 0 {
		return nil // , nil // NULL
	}
	blob, err := Crypto().Base().Encrypt(
		context.Background(), plain,
	)
	if err != nil {
		// failed to encrypt sensitive data
		// will retry on next time update ..
		args := []any{"err", err}
		path := strings.Join(tableColumn, ".")
		if path = strings.Trim(path, "."); path != "" {
			args = append(args, "column", path)
		}
		slog.Warn(
			"Failed to encrypt sensitive data ; Keeping it plain until next data update",
			args...,
		)
		blob = []byte(plain)
		// return plain, err
	}
	return blob // , nil
}

func decryptBytes(into *[]byte) sql.Scanner {
	return scanFunc(func(src any) (err error) {
		// sanitize: NULL
		(*into) = nil
		// ciphertext
		var blob []byte
		err = scanBytes(&blob)(src)
		if err != nil || len(blob) == 0 {
			return err // NULL -or- ERROR
		}
		// DECRYPT (optional)
		(*into), _, err = Crypto().Base().Decrypt(
			context.Background(), blob,
		)
		if err != nil {
			// failed to decrypt data
			(*into) = blob // ciphertext
			// err = fmt.Errorf("cryptostore/field: decrypt %s.%s row(id:%s): %w", "schema.table", "column", row.ID, err)
			// return err
			return errors.New("store.sql.convert_cipher_text")
		}
		// OK
		// (*into) == plaintext
		return nil
	})
}

func encryptText(plain string, tableColumn ...string) []byte {
	if plain == "" {
		return nil // NULL
	}
	return encryptBytes([]byte(plain), tableColumn...)
}

func decryptText(into *string) sql.Scanner {
	return scanFunc(func(src any) (err error) {
		// sanitize: NULL
		(*into) = ""
		// ciphertext
		var text []byte
		err = decryptBytes(&text).Scan(src)
		if err != nil || len(text) == 0 {
			return err // NULL -or- ERROR
		}
		// OK
		(*into) = string(text)
		return nil
	})
}

var (
	jsonNull = []byte("null")
	jsonZeroArray = []byte("[]")
	jsonZeroObject = []byte("{}")
)

// encrypts JSON column nested field(s) according to cryptostore/schema.Unit(table) declaration
func encryptJSONSchema(src any, table, column string) drv.Valuer {
	return valueFunc(func() (drv.Value, error) {
		
		if src == nil {
			return nil, nil
		}

		data, err := json.Marshal(src)
		if err != nil {
			return nil, err
		}

		for _, none := range [][]byte{
			jsonZeroObject,
			jsonZeroArray,
			jsonNull,
		} {
			if bytes.EqualFold(data, none) {
				data = nil
				break
			}
		}

		record := schema.Record{
			column: data,
		}
		crypto := Crypto().Unit(table)
		err = crypto.EncodeRecords(
			context.Background(), []schema.Record{record},
		)
		if err != nil {
			// failed to encrypt
			return nil, fmt.Errorf("cryptostore: encrypt %s.%s ; %w", table, column, err)
		}
		data = record[column].([]byte)
		return json.RawMessage(data), nil
	})
}

// decrypts JSON column nested field(s) according to cryptostore/schema.Unit(table) declaration
func decryptJSONSchema(dst any, table, column string) sql.Scanner {
	return scanFunc(func(src any) (err error) {
		
		var data []byte
		err = scanBytes(&data).Scan(src)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil // NULL
		}

		record := schema.Record{
			column: data,
		}
		crypto := Crypto().Unit(table)
		err = crypto.DecodeRecords(
			context.Background(), []schema.Record{record},
		)
		if err != nil {
			// failed to decrypt
			return fmt.Errorf("cryptostore: decrypt %s.%s ; %w", table, column, err)
		}

		data = record[column].([]byte)
		err = json.Unmarshal(data, dst)
		if err != nil {
			return fmt.Errorf("sql: convert JSON into %T", dst)
		}
		
		return nil
	})
}

// examinate JSON value structure and variadically decrypts encrypted value(s)..
func decryptJSON(dst any) sql.Scanner {
	return scanFunc(func(src any) (err error) {

		var data []byte
		err = scanBytes(&data).Scan(src)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			return nil // NULL
		}

		// 1. walk down thru JSON value.(type)
		// 2. if string("cbox:") - try to decrypt !
		data, err = schema.DecryptJSONB(
			Crypto().Base(), data,
		)
		if err != nil {
			return fmt.Errorf("cryptostore: decrypt %T ; %w", dst, err)
		}
		
		err = json.Unmarshal(data, dst)
		if err != nil {
			return fmt.Errorf("sql: convert JSON into %T", dst)
		}
		
		return nil
	})
}

type scanFunc func(src any) error

var _ sql.Scanner = scanFunc(nil)

func (scan scanFunc) Scan(src any) error {
	if scan == nil {
		return nil
	}
	return scan(src)
}

func scanBytes(dst *[]byte) scanFunc {
	return func(src any) error {
		*dst = nil // DEFAULT NULL
		if src == nil {
			return nil // NULL
		}
		switch data := src.(type) {
		case []byte:
			if data == nil {
				return nil // NULL
			}
			*dst = append(*dst, data...) // copy
		default:
			return errors.New("store.sql.convert_bytes")
			// return errors.Errorf("[database]: convert %[1]T value '%[1]v' to []byte", src)
		}
		return nil
	}
}

type valueFunc func() (drv.Value, error)

var _ drv.Valuer = valueFunc(nil)

// Value implements database/sql/driver.Valuer interface
func (eval valueFunc) Value() (drv.Value, error) {
	return eval()
}