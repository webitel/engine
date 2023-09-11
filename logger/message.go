package logger

type RequiredFields struct {
	UserId   int    `json:"userId,omitempty"`
	UserIp   string `json:"userIp,omitempty"`
	Action   string `json:"action,omitempty"`
	Date     int64  `json:"date,omitempty"`
	DomainId int64
	ObjectId int64
}

type Message struct {
	RecordsStates  map[int][]byte `json:"records,omitempty"`
	NewState       []byte         `json:"newState,omitempty"`
	RecordId       int            `json:"recordId,omitempty"`
	RequiredFields `json:"requiredFields"`
}
