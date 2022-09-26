package model

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
	"github.com/pborman/uuid"
)

const (
	LOWERCASE_LETTERS = "abcdefghijklmnopqrstuvwxyz"
	UPPERCASE_LETTERS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NUMBERS           = "0123456789"
	SYMBOLS           = " !\"\\#$%&'()*+,-./:;<=>?@[]^_`|~"
)

type StringInterface map[string]interface{}
type StringMap map[string]string
type StringArray []string
type Int64Array []int64
type Lookup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (l *Lookup) GetSafeId() *int {
	if l == nil || l.Id == 0 {

		return nil
	}

	return &l.Id
}

func LookupIds(l []*Lookup) []int {
	if l == nil {
		return nil
	}

	var res = make([]int, 0, len(l))
	for _, v := range l {
		res = append(res, v.Id)
	}

	return res
}

func (s StringMap) ToJson() string {
	if s == nil {
		return "{}"
	}

	data, _ := json.Marshal(s)
	return string(data)
}

func (s *StringMap) ToSafeJson() *string {
	var res *string
	if s == nil {
		return res
	}

	data, _ := json.Marshal(s)
	res = new(string)
	*res = string(data)
	return res
}

func (s StringInterface) ToSafeJson() *string {
	var res *string
	if s == nil {
		return res
	}

	data, _ := json.Marshal(s)
	res = new(string)
	*res = string(data)
	return res
}

func (sa StringArray) Equals(input StringArray) bool {

	if len(sa) != len(input) {
		return false
	}

	for index := range sa {

		if sa[index] != input[index] {
			return false
		}
	}

	return true
}

var translateFunc goi18n.TranslateFunc = nil

func AppErrorInit(t goi18n.TranslateFunc) {
	translateFunc = t
}

type AppError struct {
	Id            string `json:"id"`
	Message       string `json:"status"`               // Message to be display to the end user without debugging information
	DetailedError string `json:"detail"`               // Internal error string to help the developer
	RequestId     string `json:"request_id,omitempty"` // The RequestId that's also set in the header
	StatusCode    int    `json:"code,omitempty"`       // The http status code
	Where         string `json:"-"`                    // The function where it happened in the form of Struct.Func
	params        map[string]interface{}
}

func (er *AppError) Error() string {
	return er.Where + ": " + er.Message + ", " + er.DetailedError
}

func (er *AppError) Translate(T goi18n.TranslateFunc) {
	if T == nil {
		er.Message = er.Id
		return
	}

	if er.params == nil {
		er.Message = T(er.Id)
	} else {
		er.Message = T(er.Id, er.params)
	}
}

func (er *AppError) SystemMessage(T goi18n.TranslateFunc) string {
	if er.params == nil {
		return T(er.Id)
	} else {
		return T(er.Id, er.params)
	}
}

func (er *AppError) ToJson() string {
	b, _ := json.Marshal(er)
	return string(b)
}

func (er *AppError) String() string {
	if er.Id == er.Message && er.DetailedError != "" {
		return er.DetailedError
	}

	return er.Message
}

// AppErrorFromJson will decode the input and return an AppError
func AppErrorFromJson(data io.Reader) *AppError {
	str := ""
	bytes, rerr := ioutil.ReadAll(data)
	if rerr != nil {
		str = rerr.Error()
	} else {
		str = string(bytes)
	}

	decoder := json.NewDecoder(strings.NewReader(str))
	var er AppError
	err := decoder.Decode(&er)
	if err == nil {
		return &er
	} else {
		return NewAppError("AppErrorFromJson", "model.utils.decode_json.app_error", nil, "body: "+str, http.StatusInternalServerError)
	}
}

func NewAppError(where string, id string, params map[string]interface{}, details string, status int) *AppError {
	ap := &AppError{}
	ap.Id = id
	ap.params = params
	ap.Message = id
	ap.Where = where
	ap.DetailedError = details
	ap.StatusCode = status
	ap.Translate(translateFunc)
	return ap
}

var encoding = base32.NewEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

// NewId is a globally unique identifier.  It is a [A-Z0-9] string 26
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off.
func NewId() string {
	var b bytes.Buffer
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(uuid.NewRandom())
	encoder.Close()
	b.Truncate(26) // removes the '==' padding
	return b.String()
}

func NewUuid() string {
	return uuid.NewRandom().String()
}

func NewRandomString(length int) string {
	var b bytes.Buffer
	str := make([]byte, length+8)
	rand.Read(str)
	encoder := base32.NewEncoder(encoding, &b)
	encoder.Write(str)
	encoder.Close()
	b.Truncate(length) // removes the '==' padding
	return b.String()
}

func GetTime() *time.Time {
	t := time.Now()
	return &t
}

// GetMillis is a convenience method to get milliseconds since epoch.
func GetMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetMillisForTime is a convenience method to get milliseconds since epoch for provided Time.
func GetMillisForTime(thisTime time.Time) int64 {
	return thisTime.UnixNano() / int64(time.Millisecond)
}

// PadDateStringZeros is a convenience method to pad 2 digit date parts with zeros to meet ISO 8601 format
func PadDateStringZeros(dateString string) string {
	parts := strings.Split(dateString, "-")
	for index, part := range parts {
		if len(part) == 1 {
			parts[index] = "0" + part
		}
	}
	dateString = strings.Join(parts[:], "-")
	return dateString
}

// GetStartOfDayMillis is a convenience method to get milliseconds since epoch for provided date's start of day
func GetStartOfDayMillis(thisTime time.Time, timeZoneOffset int) int64 {
	localSearchTimeZone := time.FixedZone("Local Search Time Zone", timeZoneOffset)
	resultTime := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, localSearchTimeZone)
	return GetMillisForTime(resultTime)
}

// GetEndOfDayMillis is a convenience method to get milliseconds since epoch for provided date's end of day
func GetEndOfDayMillis(thisTime time.Time, timeZoneOffset int) int64 {
	localSearchTimeZone := time.FixedZone("Local Search Time Zone", timeZoneOffset)
	resultTime := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 23, 59, 59, 999999999, localSearchTimeZone)
	return GetMillisForTime(resultTime)
}

func CopyStringMap(originalMap map[string]string) map[string]string {
	copyMap := make(map[string]string)
	for k, v := range originalMap {
		copyMap[k] = v
	}
	return copyMap
}

// MapToJson converts a map to a json string
func MapToJson(objmap map[string]string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

// MapToJson converts a map to a json string
func MapBoolToJson(objmap map[string]bool) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

// MapFromJson will decode the key/value pair map
func MapFromJson(data io.Reader) map[string]string {
	decoder := json.NewDecoder(data)

	var objmap map[string]string
	if err := decoder.Decode(&objmap); err != nil {
		return make(map[string]string)
	} else {
		return objmap
	}
}

// MapFromJson will decode the key/value pair map
func MapBoolFromJson(data io.Reader) map[string]bool {
	decoder := json.NewDecoder(data)

	var objmap map[string]bool
	if err := decoder.Decode(&objmap); err != nil {
		return make(map[string]bool)
	} else {
		return objmap
	}
}

func ArrayToJson(objmap []string) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func ArrayFromJson(data io.Reader) []string {
	decoder := json.NewDecoder(data)

	var objmap []string
	if err := decoder.Decode(&objmap); err != nil {
		return make([]string, 0)
	} else {
		return objmap
	}
}

func ArrayFromInterface(data interface{}) []string {
	stringArray := []string{}

	dataArray, ok := data.([]interface{})
	if !ok {
		return stringArray
	}

	for _, v := range dataArray {
		if str, ok := v.(string); ok {
			stringArray = append(stringArray, str)
		}
	}

	return stringArray
}

func StringInterfaceToJson(objmap map[string]interface{}) string {
	b, _ := json.Marshal(objmap)
	return string(b)
}

func StringInterfaceFromJson(data io.Reader) map[string]interface{} {
	decoder := json.NewDecoder(data)

	var objmap map[string]interface{}
	if err := decoder.Decode(&objmap); err != nil {
		return make(map[string]interface{})
	} else {
		return objmap
	}
}

func StringToJson(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func StringFromJson(data io.Reader) string {
	decoder := json.NewDecoder(data)

	var s string
	if err := decoder.Decode(&s); err != nil {
		return ""
	} else {
		return s
	}
}

func IsLower(s string) bool {
	return strings.ToLower(s) == s
}

func IsValidEmail(email string) bool {
	if !IsLower(email) {
		return false
	}

	if addr, err := mail.ParseAddress(email); err != nil {
		return false
	} else if addr.Name != "" {
		// mail.ParseAddress accepts input of the form "Billy Bob <billy@example.com>" which we don't allow
		return false
	}

	return true
}

var reservedName = []string{
	"signup",
	"login",
	"admin",
	"channel",
	"post",
	"api",
	"oauth",
	"error",
	"help",
}

var validHashtag = regexp.MustCompile(`^(#\pL[\pL\d\-_.]*[\pL\d])$`)
var puncStart = regexp.MustCompile(`^[^\pL\d\s#]+`)
var hashtagStart = regexp.MustCompile(`^#{2,}`)
var puncEnd = regexp.MustCompile(`[^\pL\d\s]+$`)

func ParseHashtags(text string) (string, string) {
	words := strings.Fields(text)

	hashtagString := ""
	plainString := ""
	for _, word := range words {
		// trim off surrounding punctuation
		word = puncStart.ReplaceAllString(word, "")
		word = puncEnd.ReplaceAllString(word, "")

		// and remove extra pound #s
		word = hashtagStart.ReplaceAllString(word, "#")

		if validHashtag.MatchString(word) {
			hashtagString += " " + word
		} else {
			plainString += " " + word
		}
	}

	if len(hashtagString) > 1000 {
		hashtagString = hashtagString[:999]
		lastSpace := strings.LastIndex(hashtagString, " ")
		if lastSpace > -1 {
			hashtagString = hashtagString[:lastSpace]
		} else {
			hashtagString = ""
		}
	}

	return strings.TrimSpace(hashtagString), strings.TrimSpace(plainString)
}

func IsValidHttpUrl(rawUrl string) bool {
	if strings.Index(rawUrl, "http://") != 0 && strings.Index(rawUrl, "https://") != 0 {
		return false
	}

	if _, err := url.ParseRequestURI(rawUrl); err != nil {
		return false
	}

	return true
}

func IsValidId(value string) bool {
	if len(value) != 26 {
		return false
	}

	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

func InterfaceToMapString(val interface{}) map[string]interface{} {
	var res map[string]interface{}
	data, _ := json.Marshal(val)
	json.Unmarshal(data, &res)

	return res
}

func HideString(str string, minimumNumberMaskLen, pref, suff int) string {
	l := len(str)
	h := pref + suff

	if l > minimumNumberMaskLen && l > (h+1) && minimumNumberMaskLen >= (h+1) {
		return str[:pref] + string(makeX(l-h)) + str[l-suff:l]
	}

	return str
}

func makeX(cnt int) []rune {
	r := make([]rune, cnt, cnt)
	for i := 0; i < cnt; i++ {
		r[i] = '*'
	}

	return r
}
