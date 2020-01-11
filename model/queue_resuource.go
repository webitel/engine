package model

type QueueResourceGroup struct {
	Id            int64  `json:"id" db:"id"`
	QueueId       int64  `json:"queue_id" db:"queue_id"`
	ResourceGroup Lookup `json:"resource_group" db:"resource_group"`
}

func (q *QueueResourceGroup) IsValid() *AppError {
	//FIXME
	return nil
}
