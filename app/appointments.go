package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/utils"
	"golang.org/x/sync/singleflight"
	"io"
	"net/http"
	"strconv"
)

const (
	sizeCacheAppointments = 10000
)

var (
	cacheAppointments       utils.ObjectCache
	cacheAppointmentDate    utils.ObjectCache
	appointmentGroupRequest singleflight.Group
	cipherKey               = []byte("asuperstrong32bitpasswordgohere!") // config
)

func init() {
	cacheAppointments = utils.NewLruWithParams(sizeCacheAppointments, "Appointment", 60, "")
	cacheAppointmentDate = utils.NewLruWithParams(sizeCacheAppointments, "List appointment date", 60, "")
}

func (app *App) GetAppointment(key string) (*model.Appointment, *model.AppError) {
	if a, ok := cacheAppointments.Get(key); ok {
		return a.(*model.Appointment), nil
	}

	memberId, appErr := decryptMemberId(key)
	if appErr != nil {
		return nil, appErr
	}

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("member-%d", memberId), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointment(memberId)
		if err != nil {
			return nil, err
		}

		a.Computed = (&model.AppointmentResponse{
			Timezone:    a.Timezone,
			Type:        "appointment",
			List:        nil,
			Appointment: a,
		}).ToJSON()

		return a, nil
	})

	if err != nil {
		switch err.(type) {
		case *model.AppError:
			return nil, err.(*model.AppError)
		default:
			return nil, model.NewAppError("App", "app.appointment.get", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if !shared {
		cacheAppointments.AddWithDefaultExpires(key, res)
	}

	return res.(*model.Appointment), nil
}

func (app *App) AppointmentWidget(widgetUri string) (*model.AppointmentWidget, *model.AppError) {
	if a, ok := cacheAppointmentDate.Get(widgetUri); ok {
		return a.(*model.AppointmentWidget), nil
	}

	return app.appointmentWidget(widgetUri)
}

func (app *App) appointmentWidget(widgetUri string) (*model.AppointmentWidget, *model.AppError) {

	res, err, shared := appointmentGroupRequest.Do(fmt.Sprintf("list-%s", widgetUri), func() (interface{}, error) {
		a, err := app.Store.Member().GetAppointmentWidget(widgetUri)
		if err != nil {
			return nil, err
		}
		a.ComputedList = (&model.AppointmentResponse{
			Timezone: a.Profile.Timezone,
			Type:     "list",
			List:     a.List,
		}).ToJSON()

		return a, nil
	})

	if err != nil {
		switch err.(type) {
		case *model.AppError:
			return nil, err.(*model.AppError)
		default:
			return nil, model.NewAppError("App", "app.appointment.list", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if !shared {
		cacheAppointmentDate.AddWithDefaultExpires(widgetUri, res)
	}

	return res.(*model.AppointmentWidget), nil
}

func (app *App) CreateAppointment(widget *model.AppointmentWidget, appointment *model.Appointment) (*model.Appointment, *model.AppError) {
	var err *model.AppError
	if !widget.ValidAppointment(appointment) {
		return nil, model.NewAppError("CreateAppointment", "appointment.valid.date", nil, "No slot", http.StatusBadRequest)
	}

	appointment, err = app.Store.Member().CreateAppointment(&widget.Profile, appointment)
	if err != nil {
		return nil, err
	}

	appointment.Key, err = encrypt(cipherKey, []byte(fmt.Sprintf("%v", appointment.Id)))
	if err != nil {
		return nil, err
	}

	appointment.Timezone = widget.Profile.Timezone
	appointment.Computed = (&model.AppointmentResponse{
		Timezone:    widget.Profile.Timezone,
		Type:        "appointment",
		List:        nil,
		Appointment: appointment,
	}).ToJSON()

	// reset list ?
	app.appointmentWidget(widget.Profile.Uri)

	return appointment, nil
}

func (app *App) CancelAppointment(widget *model.AppointmentWidget, key string) (*model.Appointment, *model.AppError) {
	appointment, err := app.GetAppointment(key)
	if err != nil {
		return nil, err
	}

	if err = app.Store.Member().CancelAppointment(appointment.Id, "cancel"); err != nil {
		return nil, err
	}

	// reset list ?
	app.appointmentWidget(widget.Profile.Uri)

	return appointment, nil
}

func decryptMemberId(key string) (int64, *model.AppError) {
	val, appErr := decrypt(cipherKey, key)
	if appErr != nil {
		return 0, appErr
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, model.NewAppError("App", "app.appointment.decrypt_member", nil, err.Error(), http.StatusBadRequest)
	}

	return int64(id), nil
}

func encrypt(key []byte, plainText []byte) (string, *model.AppError) {
	//Create a new AES cipher using the key
	block, err := aes.NewCipher(key)

	//IF NewCipher failed, exit:
	if err != nil {
		return "", model.NewAppError("App", "app.appointment.encrypt", nil, err.Error(), http.StatusBadRequest)
	}

	//Make the cipher text a byte array of size BlockSize + the length of the message
	cipherText := make([]byte, aes.BlockSize+len(plainText))

	//iv is the ciphertext up to the blocksize (16)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", model.NewAppError("App", "app.appointment.encrypt", nil, err.Error(), http.StatusBadRequest)
	}

	//Encrypt the data:
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//Return string encoded in base64
	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

func decrypt(key []byte, secure string) (string, *model.AppError) {
	//Remove base64 encoding:
	cipherText, err := base64.RawStdEncoding.DecodeString(secure)

	//IF DecodeString failed, exit:
	if err != nil {
		return "", model.NewAppError("App", "app.appointment.decrypt", nil, err.Error(), http.StatusBadRequest)
	}

	//Create a new AES cipher with the key and encrypted message
	var block cipher.Block
	block, err = aes.NewCipher(key)

	//IF NewCipher failed, exit:
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
