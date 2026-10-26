package sqlstore

import (
	"reflect"
	"testing"

	cryptoschema "github.com/webitel/crypto/cryptostore/schema"
	wcrypto "github.com/webitel/webitel-cryptoctl/schema"
)

// TestCryptoBaselineIsCallCenter: the encryption baseline is the shared call_center
// schema from webitel-cryptoctl (also used by flow_manager and cryptoctl), not a local copy.
func TestCryptoBaselineIsCallCenter(t *testing.T) {
	t.Setenv("CRYPTO_CIPHER_KEYRING", "correct-horse-battery-staple-001!")
	if err := CryptoInit(); err != nil {
		t.Fatal(err)
	}
	if got := cryptoschema.Baseline(); !reflect.DeepEqual(got, wcrypto.CallCenter) {
		t.Fatalf("baseline = %+v, want webitel-cryptoctl CallCenter %+v", got, wcrypto.CallCenter)
	}
	if Crypto().Unit(schemaTableEmailAccount) == nil {
		t.Fatalf("codec has no unit %s", schemaTableEmailAccount)
	}
}
