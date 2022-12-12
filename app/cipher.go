package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/webitel/engine/model"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (app *App) setupCipher() *model.AppError {
	var privPem *pem.Block
	cipherKey, err := os.ReadFile(app.config.PresignedCert)
	if err != nil {
		return model.NewAppError("App", "app.cipher_key.load.app_error", nil,
			err.Error(), http.StatusInternalServerError)
	}

	privPem, cipherKey = pem.Decode(cipherKey)
	cipherKey = privPem.Bytes[0:32]

	//Create a new AES cipher using the key
	app.cipherBlock, err = aes.NewCipher(cipherKey)
	if err != nil {
		return model.NewAppError("App", "app.cipher_key.create.app_error", nil,
			err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (app *App) DecryptId(key string) (int64, *model.AppError) {
	val, appErr := decrypt(app.cipherBlock, key)
	if appErr != nil {
		return 0, appErr
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, model.NewAppError("App", "app.appointment.decrypt_member", nil, err.Error(), http.StatusBadRequest)
	}

	return int64(id), nil
}

func (app *App) EncryptId(id int64) (string, *model.AppError) {
	return encrypt(app.cipherBlock, []byte(fmt.Sprintf("%v", id)))
}

func encrypt(block cipher.Block, plainText []byte) (string, *model.AppError) {

	//Make the cipher text a byte array of size BlockSize + the length of the message
	cipherText := make([]byte, aes.BlockSize+len(plainText))

	//iv is the ciphertext up to the blocksize (16)
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", model.NewAppError("App", "app.appointment.encrypt", nil, err.Error(), http.StatusBadRequest)
	}

	//Encrypt the data:
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//Return string encoded in base64
	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

func decrypt(block cipher.Block, secure string) (string, *model.AppError) {
	//Remove base64 encoding:
	cipherText, err := base64.RawStdEncoding.DecodeString(secure)

	//IF DecodeString failed, exit:
	if err != nil {
		return "", model.NewAppError("App", "app.appointment.decrypt", nil, err.Error(), http.StatusBadRequest)
	}

	//IF the length of the cipherText is less than 16 Bytes:
	if len(cipherText) < aes.BlockSize {
		return "", model.NewAppError("App", "app.appointment.decrypt", nil, "Ciphertext block size is too short!", http.StatusBadRequest)
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	//Decrypt the message
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
