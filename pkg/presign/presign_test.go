package presign_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/webitel/engine/pkg/presign"
)

func TestMain(test *testing.M) {

	// export := os.Setenv

	// export("WBTL_CRYPTO_DIR", "") // modern
	// export(testEnvPrivateKey, "") // legacy

	test.Run()
}

const testEnvPrivateKey = "PRESIGN_KEYFILE"

func legacy(test *testing.T) (presign.PreSign) {
	test.Helper()

	keyfile := os.Getenv(testEnvPrivateKey)
	if keyfile == "" {
		test.Skipf("%s = ?", testEnvPrivateKey)
		return nil
	}
	
	impl, err := presign.NewPrivateKey(keyfile)
	if err != nil {
		test.Errorf("legacy: configuration failed ;  %v", err)
	}
	return impl
}

func modern(test *testing.T) (presign.PreSign) {
	test.Helper()

	cbox, err := presign.NewCryptoBox(nil)
	if err != nil {
		test.Errorf("modern: configuration failed ; %v", err)
	}
	return cbox
}

func latest(test *testing.T) (presign.PreSign) {
	test.Helper()

	impl, err := presign.NewPreSigned(
		os.Getenv(testEnvPrivateKey),
	)
	if err != nil {
		test.Errorf("latest: configuration failed ; %v", err)
	}
	return impl
}

func TestDummy(test *testing.T) {
	v0, v1 := legacy(test), modern(test)

	const inputText = "https://example.org/files/2026/08/03/image_20260803_153728.jpg"

	sign0, err := v0.Generate([]byte(inputText))
	if err != nil {
		test.Errorf("pkey.Generate: %v", err)
	}

	sign1, err := v1.Generate([]byte(inputText))
	if err != nil {
		test.Errorf("cbox.Generate: %v", err)
	}

	blob0, err := v0.EncryptBytes([]byte(inputText))
	if err != nil {
		test.Errorf("pkey.EncryptBytes: %v", err)
	}

	blob1, err := v1.EncryptBytes([]byte(inputText))
	if err != nil {
		test.Errorf("cbox.EncryptBytes: %v", err)
	}

	test.Logf("[input]: %s", inputText)
	test.Logf("[pkey.Generate]: %s", sign0)
	test.Logf("[cbox.Generate]: %s", sign1)
	test.Logf("[pkey.EncryptBytes]: %s", blob0)
	test.Logf("[cbox.EncryptBytes]: %s", blob1)

}

func TestSignatureRoundTrip(test *testing.T) {

	tests := []struct{
		name string
		build func(*testing.T) presign.PreSign
	}{
		{"pkey", legacy},
		{"cbox", modern},
	}

	const inputText = "https://example.org/files/2026/08/03/image_20260803_153728.jpg"
	test.Logf("[plaintext]: %s", inputText)
	
	for _, tcase := range tests {
		test.Run(tcase.name, func(t *testing.T) {
			
			enc := tcase.build(t)
			
			sign, err := enc.Generate([]byte(inputText))
			if err != nil {
				test.Errorf("Generate: %v", err)
			}
			test.Logf("[Signature]: %s", sign)
			if !enc.Valid(inputText, sign) {
				test.Errorf("Verify: %v", err)
			}
			// // make invalid
			// sign[len(sign)-1] = '0'
		})
	}

}

func TestEncryptionRoundTrip(test *testing.T) {

	tests := []struct{
		name string
		build func(*testing.T) presign.PreSign
	}{
		{"pkey", legacy},
		{"cbox", modern},
	}

	const plaintext = "https://example.org/files/2026/08/03/image_20260803_153728.jpg"
	test.Logf("[plaintext]: %s", plaintext)

	src := []byte(plaintext)
	for _, tcase := range tests {
		test.Run(tcase.name, func(t *testing.T) {

			enc := tcase.build(t)
			
			blob, err := enc.EncryptBytes(src)
			if err != nil {
				t.Errorf("EncryptBytes: %v", err)
			}
			t.Logf("EncryptBytes: %s", blob)
			dst, err := enc.DecryptBytes(blob)
			if err != nil {
				t.Errorf("DecryptBytes: %v", err)
			}
			
			if !bytes.Equal(dst, src) {
				test.Errorf("DecryptBytes: unexpected data")
			}
			// // make invalid
			// sign[len(sign)-1] = '0'
		})
	}

}


func TestNumberEncryptionRoundTrip(test *testing.T) {

	tests := []struct{
		name string
		build func(*testing.T) presign.PreSign
	}{
		{"pkey", legacy},
		{"cbox", modern},
	}

	const oid int64 = 56843
	test.Logf("[int64]: %d", oid)

	for _, tcase := range tests {
		test.Run(tcase.name, func(t *testing.T) {

			enc := tcase.build(t)
			
			blob, err := enc.EncryptId(oid)
			if err != nil {
				t.Errorf("EncryptId: %v", err)
			}
			t.Logf("EncryptId: %s", blob)
			got, err := enc.DecryptId(blob)
			if err != nil {
				t.Errorf("DecryptId: %v", err)
			}
			
			if got != oid {
				test.Errorf("decrypted number MUST be equal")
			}
			// // make invalid
			// sign[len(sign)-1] = '0'
		})
	}

}

func _TestDecryptBytes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		blob    []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		// {"0", []byte("Em2vIKUfQu0DCuSbn3NxbcZui/Csonr3uKCwUKe6uZc2d8wgtU5FG6M9dmQHYyHZLRQ"), []byte(""), false},
		// {"1", []byte("fUoW98cD2eb7mZk+MBME/jrClhFBe9jhSA"), []byte(""), false},
		// {"2", []byte("Q+gfDweWVnpBk/EyFF6IGo2t5yWLqPQ"), []byte(""), false},
		// {"3", []byte("xseElYDZHJtIDTMmkmGT/qHrn90nwvo"), []byte(""), false},
		// {"4", []byte("2kGCoOQRNmCbMNP0Wu2ARIHMvM0R2gUZIA"), []byte(""), false},
		// {"5", []byte("qNSzUoo+orLnc541Y+BVZlbTx9gN"), []byte(""), false},
		// {"6", []byte("QLBp7rQ4pRBq3am6nHUKtCpBMP7R"), []byte(""), false},
		{"7", []byte("dQzXsQVU0tMuwaWaCcgCYgrfmjGKrsy3"), []byte(`"qwerty"`), false},
	}

	enc := latest(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := enc.DecryptBytes(tt.blob)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DecryptBytes() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DecryptBytes() succeeded unexpectedly")
			}
			t.Logf("DecryptBytes() = %s", got)
			// TODO: update the condition below to compare got with tt.want.
			if !bytes.Equal(got, tt.want) {
				t.Errorf("DecryptBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}