package sqlstore

import (
	"fmt"
	"sync"

	"github.com/webitel/crypto/cryptostore/schema"
)

// CryptoInit configures cryptostore/schema.Codec plugin from environment
func CryptoInit() error {
	crypto.init.Do(cryptoInit)
	return crypto.err
}

// guard: crypto.init.(sync.Once)
func cryptoInit() {
	// cryptostore/schema.(BASELINE)
	schema.Register(&crypto.schema)
	// plugin: module environment configuration
	crypto.codec, crypto.err = schema.NewCodec(
		schema.DefaultOptions(),
	)
	if crypto.err == nil {
		// schema MAY define field(s) with the "search" option enabled
		// this require WBTL_CRYPTO_SEARCH_{KERING|KEYFILE} to be specified
		crypto.err = crypto.codec.RequireIndex()
	}
	if crypto.err != nil {
		// wrap up general error details
		crypto.err = fmt.Errorf("crypto: configuration ; %w", crypto.err)
	}
}

// Crypto schema.Codec for data encryption
func Crypto() *schema.Codec {
	err := CryptoInit() // lazy: init
	if err != nil {
		panic(err) // configuration failed
	}
	return crypto.codec
}

const (
	schemaTableEmailAccount = "call_center.cc_email_profile"
)

// module: cryptostore/schema
var crypto = struct {
	// baseline (mandatory) schema
	schema schema.Config
	codec *schema.Codec
	init sync.Once
	err error
} {

	schema: schema.Config{
		Version: 1,
		Units: map[string]*schema.Unit{
			schemaTableEmailAccount: {
				Fields: map[string]*schema.FieldPolicy{
					"password": {},
					"params": {Nested: []schema.FieldNested{
						{Path: []string{"$.oauth2.client_secret"}},
					}},
					"token": {Nested: []schema.FieldNested{
						{Path: []string{"access_token"}},
						{Path: []string{"refresh_token"}},
					}},
				},
			},
		},
	},

}
