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
	Id       int64  `json:"id" db:"id"`
	QueueId  int64  `json:"queue_id" db:"queue_id"`
	Bucket   Lookup `json:"bucket" db:"bucket"`
	Ratio    int    `json:"ratio" db:"ratio"`
	Priority int    `json:"priority" db:"priority"`
	Disabled bool   `json:"disabled" db:"disabled"`
}

type SearchQueueBucket struct {
	ListRequest
	Ids []uint32
}

type QueueBucketPatch struct {
	Bucket   *Lookup `json:"bucket" db:"bucket"`
	Ratio    *int    `json:"ratio" db:"ratio"`
	Priority *int    `json:"priority" db:"priority"`
	Disabled *bool   `json:"disabled" db:"disabled"`
}

func (q *QueueBucket) Patch(patch *QueueBucketPatch) {
	if patch.Bucket != nil {
		q.Bucket = *patch.Bucket
	}
	if patch.Priority != nil {
		q.Priority = *patch.Priority
	}
	if patch.Disabled != nil {
		q.Disabled = *patch.Disabled
	}
	if patch.Ratio != nil {
		q.Ratio = *patch.Ratio
	}
}

func (QueueBucket) DefaultOrder() string {
	return "bucket_name"
}

func (a QueueBucket) AllowFields() []string {
	return []string{"id", "ratio", "bucket", "queue_id", "bucket_id", "domain_id", "bucket_name", "disabled", "priority"}
}

func (a QueueBucket) DefaultFields() []string {
	return []string{"id", "ratio", "bucket", "disabled", "priority"}
}

func (a QueueBucket) EntityName() string {
	return "cc_bucket_in_queue_view"
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

func (b *Bucket) IsValid() AppError {
	//FIXME
	return nil
}

func (q *QueueBucket) IsValid() AppError {
	//FIXME
	return nil
}
