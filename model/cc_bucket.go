package model

type Bucket struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type SearchBucket struct {
	ListRequest
}

type QueueBucket struct {
	Id      int64  `json:"id" db:"id"`
	QueueId int64  `json:"queue_id" db:"queue_id"`
	Bucket  Lookup `json:"bucket" db:"bucket"`
	Ratio   int    `json:"ratio" db:"ratio"`
}

type SearchQueueBucket struct {
	ListRequest
}

func (b *Bucket) IsValid() *AppError {
	//FIXME
	return nil
}

func (q *QueueBucket) IsValid() *AppError {
	//FIXME
	return nil
}
