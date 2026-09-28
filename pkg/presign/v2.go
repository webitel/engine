package presign

import "cmp"

// NewPreSigned builds PreSign implementation with cryptobox preferred algorithm.
// Also supports legacy PrivateKey mechanism for signature verification & decryption.
// Use [NewPrivateKey] constructor for previous implementation only.
func NewPreSigned(pemLocation string) (PreSign, error) {
	// MUST: v1. MODERN
	var modern Crypto
	err := modern.init()
	if err != nil {
		// MUST. -but- failed
		return nil, err
	}
	// WITH: v0. LEGACY ?
	if pemLocation == "" {
		// ONLY: v1 (modern) support
		return preferred{modern}, nil
	}
	// MUST v0 (legacy) support
	legacy, err := NewPrivateKey(pemLocation)
	if err != nil {
		// bad configuration
		return nil, err
	}
	// WITH: v0 (legacy) support
	return preferred{modern, legacy}, nil
}

// .well-known & supported
// [0]  - encoding ; preferred
// [1:] - decoding ; supported
type preferred []PreSign

func (x preferred) latest() PreSign {
	return x[0]
}

var _ PreSign = (preferred)(nil)

func (x preferred) Generate(data []byte) (string, error) {
	return x.latest().Generate(data)
}

func (x preferred) Valid(plaintext string, signature string) bool {
	for _, spec := range x {
		if spec.Valid(plaintext, signature) {
			return true // break
		}
	}
	return false
}

func (x preferred) EncryptId(id int64) (string, error) {
	return x.latest().EncryptId(id)
}

func (x preferred) DecryptId(key string) (int64, error) {
	var err error // remember: first (latest)
	for _, sup := range x {
		oid, res := sup.DecryptId(key)
		if res == nil {
			// success
			return oid, nil
		}
		// failure
		err = cmp.Or(err, res) // first
	}
	return 0, err
}

func (x preferred) EncryptBytes(data []byte) ([]byte, error) {
	return x.latest().EncryptBytes(data)
}

func (x preferred) DecryptBytes(text []byte) ([]byte, error) {
	var err error // remember: first (latest)
	for _, sup := range x {
		data, res := sup.DecryptBytes(text)
		if res == nil {
			// success
			return data, nil
		}
		// failure
		err = cmp.Or(err, res) // first
	}
	return nil, err
}

