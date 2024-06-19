package model

type SendPush struct {
	Android    []string
	Apple      []string
	Data       map[string]string
	Expiration int64
	Priority   int32
}
