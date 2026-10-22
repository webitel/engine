package presign

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/webitel/crypto/cryptobox"
)

// modern cryptobox.Cipher implementation
type Crypto struct {
	box cryptobox.Cipher
}

// NewCryptoBox implements modern ciphertext encryption strategy
func NewCryptoBox(box cryptobox.Cipher) (PreSign, error) {
	cbox := Crypto{box}
	err := cbox.init()
	if err != nil {
		return nil, err
	}
	return cbox, nil
}

// lazy init
func (c *Crypto) init() (err error) {
	if c.box != nil {
		return nil // already
	}
	// Load environment configuration
	c.box, err = cryptobox.Default()
	if err != nil {
		return err
	}
	// OK
	return nil
}

var _ PreSign = Crypto{}

var hashSum = hash

func (c Crypto) Generate(data []byte) (string, error) {
	
	err := c.init()
	if err != nil {
		return "", err
	}

	sign := hashSum(data)
	sign, err = c.box.Encrypt(
		context.Background(), sign,
	)
	if err != nil {
		// failed to encrypt hash of the given data message
		return "", err
	}

	return hex.EncodeToString(sign), nil
}

func (c Crypto) Valid(plaintext string, signature string) bool {

	sign, err := hex.DecodeString(signature)
	if err != nil {
		// failed to decode signature
		return false
	}

	err = c.init()
	if err != nil {
		return false
	}

	// v2
	sign, err = c.box.Decrypt(
		context.Background(), sign,
	)
	if err != nil {
		// failed to decrypt signature
		return false
	}

	// verify
	want := hashSum([]byte(plaintext))
	return bytes.Equal(want, sign)
}

func (c Crypto) EncryptId(id int64) (string, error) {
	text, err := c.encryptText(
		binary.AppendVarint(nil, id),
	)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func (c Crypto) DecryptId(key string) (int64, error) {
	// v2 encryption
	data, err := c.decryptText([]byte(key))
	if err != nil {
		return 0, err
	}
	num, n := binary.Varint(data)
	// if n <= 0 {
	// 	// invalid input ; not integer encrypted
	// 	return 0, strconv.ErrSyntax
	// }
	if n != len(data) {
		// read too short
		return 0, strconv.ErrSyntax
	}
	return num, nil
}

func (c Crypto) EncryptBytes(v []byte) ([]byte, error) {
	// from business logic: output supposed to be the cipher TEXT (printable, NOT raw bytes)
	return c.encryptText(v)
}

func (c Crypto) DecryptBytes(v []byte) ([]byte, error) {
	// expect TEXT bytes ; see: c.EncryptBytes()
	return c.decryptText(v)
}

const cipherTag = ".c1"
var cipherText = base64.RawURLEncoding

func (c *Crypto) encryptText(data []byte) (text []byte, err error) {
	
	err = c.init()
	if err != nil {
		return nil, err
	}
	
	blob, err := c.box.Encrypt(
		context.Background(), data,
	)
	if err != nil {
		// failed to encrypt sensitive data
		return nil, err
	}
	
	// v2
	// return cipherTag + cipherText.EncodeToString(blob), nil
	return cipherText.AppendEncode([]byte(cipherTag), blob), nil
}

func (c *Crypto) decryptText(text []byte) (data []byte, err error) {
	
	text, v2 := bytes.CutPrefix(text, []byte(cipherTag))
	if !v2 {
		return nil, fmt.Errorf("presign: invalid syntax")
	}

	blob, err := cipherText.AppendDecode(nil, text)
	if err != nil {
		// failed to decode ciphertext
		return nil, err
	}

	err = c.init()
	if err != nil {
		return nil, err
	}

	data, err = c.box.Decrypt(
		context.Background(), blob,
	)
	if err != nil {
		// failed to decrypt cipherdata
		return nil, err
	}
	
	// OK
	return data, nil
}

