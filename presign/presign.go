package presign

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

//openssl genrsa -out key.pem 512

type PreSign interface {
	Generate(data []byte) (string, error)
	Valid(plaintext, signature string) bool
}

type preSign struct {
	privateKey *rsa.PrivateKey
}

func hash(msg []byte) []byte {
	sh := crypto.SHA1.New()
	sh.Write(msg)
	hash := sh.Sum(nil)
	return hash
}

func NewPreSigned(pemLocation string) (PreSign, error) {
	cert, err := ioutil.ReadFile(pemLocation)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, errors.New("decode certificate")
	}

	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &preSign{
		privateKey: pkey,
	}, nil
}

func (p *preSign) Generate(message []byte) (string, error) {
	hash := hash(message)
	bytes, err := rsa.SignPKCS1v15(rand.Reader, p.privateKey, crypto.SHA1, hash)
	if err != nil {
		panic(err)
	}
	signature := hex.EncodeToString(bytes)

	return signature, nil
}

func (p *preSign) Valid(plaintext, signature string) bool {
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	hashed := hash([]byte(plaintext))

	err = rsa.VerifyPKCS1v15(&p.privateKey.PublicKey, crypto.SHA1, hashed[:], sig)
	if err != nil {
		return false
	}

	return true
}
