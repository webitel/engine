package model

type Bucket struct {
	Id          int64  `json:"id" db:"id"`
	DomainId    int64  `json:"domain_id" db:"domain_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type QueueBucket struct {
	Id      int64  `json:"id" db:"id"`
	QueueId int64  `json:"queue_id" db:"queue_id"`
	Bucket  Lookup `json:"bucket" db:"bucket"`
	Ratio   int    `json:"ratio" db:"ratio"`
}

func (b *Bucket) IsValid() *AppError {
	//FIXME
	return nil
}

func (q *QueueBucket) IsValid() *AppError {
	//FIXME
	return nil
}
