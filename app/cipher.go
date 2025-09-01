package app

import (
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/presign"
)

func (app *App) setupCipher() model.AppError {
	var err error
	app.cipher, err = presign.NewPreSigned(app.config.PresignedCert)
	if err != nil {
		return model.NewInternalError("app.cipher_key.create.app_error", err.Error())
	}

	return nil
}

func (app *App) DecryptId(key string) (int64, model.AppError) {
	id, err := app.cipher.DecryptId(key)
	if err != nil {
		return 0, model.NewBadRequestError("app.appointment.decrypt_id", err.Error())
	}

	return id, nil
}

func (app *App) EncryptId(id int64) (string, model.AppError) {
	v, err := app.cipher.EncryptId(id)
	if err != nil {
		return "", model.NewBadRequestError("app.appointment.encrypt_id", err.Error())
	}
	return v, nil
}

func (app *App) EncryptBytes(v []byte) ([]byte, model.AppError) {
	v, err := app.cipher.EncryptBytes(v)
	if err != nil {
		return nil, model.NewBadRequestError("app.appointment.encrypt_bytes", err.Error())
	}

	return v, nil
}

func (app *App) DecryptBytes(v []byte) ([]byte, model.AppError) {
	v, err := app.cipher.DecryptBytes(v)
	if err != nil {
		return nil, model.NewBadRequestError("app.appointment.decrypt_bytes", err.Error())
	}

	return v, nil
}
