package model

type Filter struct {
	Column         string
	Value          any
	ComparisonType Comparison
}

type FilterBunch struct {
	Bunch []*Filter
	ConnectionType
}

type FilterArray struct {
	Filters    []*FilterBunch
	Connection ConnectionType
}

type FilterNode struct {
	Nodes      []any
	Connection ConnectionType
}

type Comparison int64

const (
	Equal Comparison = iota
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	NotEqual
	Like
	ILike
)

type ConnectionType int64

const (
	AND ConnectionType = 0
	OR  ConnectionType = 1
)
