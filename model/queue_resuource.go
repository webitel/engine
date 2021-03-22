package model

type QueueResourceGroup struct {
	Id            int64  `json:"id" db:"id"`
	QueueId       int64  `json:"queue_id" db:"queue_id"`
	ResourceGroup Lookup `json:"resource_group" db:"resource_group"`
}

type SearchQueueResourceGroup struct {
	ListRequest
	Ids []uint32
}

func (QueueResourceGroup) DefaultOrder() string {
	return "resource_group_name"
}

func (a QueueResourceGroup) AllowFields() []string {
	return []string{"id", "resource_group", "queue_id", "resource_group_name", "domain_id"}
}

func (a QueueResourceGroup) DefaultFields() []string {
	return []string{"id", "resource_group"}
}

func (a QueueResourceGroup) EntityName() string {
	return "cc_queue_resource_view"
}

func (q *QueueResourceGroup) IsValid() *AppError {
	//FIXME
	return nil
}
