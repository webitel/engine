package store

import (
	"context"
)

type Option func(q Query) Query

type Query struct {
	ctx    context.Context
	filter string
	Q      string
	Limit  int
	Offset int
}

func (q *Query) Build() string {
	return q.Q
}

func Select(opts ...Option) Query {
	q := Query{}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Offset(offset int) Option {
	return func(q Query) Query {
		q.Offset = offset
		return q
	}
}

func Limit(limit int) Option {
	return func(q Query) Query {
		q.Limit = limit
		return q
	}
}

func Filter(filter string) Option {
	return func(q Query) Query {
		q.filter = filter
		return q
	}
}
