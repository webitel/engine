package model

type Bucket struct {
	DomainRecord
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

type SearchBucket struct {
	ListRequest
	Ids []uint32
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

func (Bucket) DefaultOrder() string {
	return "id"
}

func (a Bucket) AllowFields() []string {
	return []string{"id", "domain_id", "name", "description"}
}

func (a Bucket) DefaultFields() []string {
	return []string{"id", "name", "description"}
}

func (a Bucket) EntityName() string {
	return "cc_bucket_view"
}

func (b *Bucket) IsValid() *AppError {
	//FIXME
	return nil
}

func (q *QueueBucket) IsValid() *AppError {
	//FIXME
	return nil
}
