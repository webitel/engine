package model

type QueueRouting struct {
	Id       int64  `json:"id" db:"id"`
	QueueId  int64  `json:"queue_id" db:"queue_id"`
	Pattern  string `json:"pattern" db:"pattern"`
	Priority int    `json:"priority" db:"priority"`
	Disabled bool   `json:"disabled" db:"disabled"`
}

func (qr *QueueRouting) IsValid() *AppError {
	//FIXME
	return nil
}
