package model

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	PAGE_DEFAULT     = 0
	PER_PAGE_DEFAULT = 40
	PER_PAGE_MAXIMUM = 20000
)

type ListRequest struct {
	Q        string
	Page     int
	PerPage  int
	DomainId int64
	endList  bool

	Fields []string
	Sort   string
}

type FilterBetween struct {
	From int64
	To   int64
}

func GetBetweenFromTime(src *FilterBetween) *time.Time {
	if src == nil || src.From == 0 {
		return nil
	}
	t := time.Unix(0, src.From*int64(time.Millisecond))
	return &t
}

func GetBetweenToTime(src *FilterBetween) *time.Time {
	if src == nil || src.To == 0 {
		return nil
	}
	t := time.Unix(0, src.To*int64(time.Millisecond))
	return &t
}

func GetBetweenFrom(src *FilterBetween) *int64 {
	if src != nil && src.From > 0 {
		return &src.From
	}

	return nil
}

func GetBetweenTo(src *FilterBetween) *int64 {
	if src != nil && src.To > 0 {
		return &src.To
	}

	return nil
}

func (l *ListRequest) RemoveLastElemIfNeed(slicePtr interface{}) {
	s := reflect.ValueOf(slicePtr)
	if s.Kind() != reflect.Ptr || s.Type().Elem().Kind() != reflect.Slice {
		panic(fmt.Errorf("first argument to Remove must be pointer to slice, not %T", slicePtr))
	}
	if s.IsNil() {
		return
	}

	itr := s.Elem()

	length := itr.Len()
	l.endList = length <= l.GetLimit()-1

	if l.endList {
		return
	}

	newSlice := reflect.MakeSlice(itr.Type(), length-1, length-1)
	reflect.Copy(newSlice.Slice(0, newSlice.Len()), itr.Slice(0, length-1))
	s.Elem().Set(newSlice)
}

func (l *ListRequest) EndOfList() bool {
	return l.endList
}

func (l *ListRequest) GetQ() *string {
	if l.Q != "" {
		return NewString(strings.Replace(l.Q, "*", "%", -1))
	}

	return nil
}

func (l *ListRequest) GetRegExpQ() *string {
	return GetRegExpQ(l.Q)
}

func GetRegExpQ(q string) *string {
	if q != "" {
		if q[0] == '+' {
			q = "\\" + q
		}
		return NewString(strings.Replace(q, "*", "%", -1))
	}

	return nil
}

func (l *ListRequest) GetLimit() int {
	l.valid()
	return l.PerPage + 1 //FIXME for next page...
}

func (l *ListRequest) GetOffset() int {
	l.valid()
	if l.Page <= 1 {
		return 0
	}
	return l.PerPage * (l.Page - 1)
}

func (l *ListRequest) valid() {
	if l.Page < 0 {
		l.Page = PAGE_DEFAULT
	}

	if l.PerPage < 1 || l.PerPage > PER_PAGE_MAXIMUM {
		l.PerPage = PER_PAGE_DEFAULT
	}
}

func GetLimit(lim int32) int32 {
	if lim == 0 {
		return PER_PAGE_DEFAULT
	}

	return lim
}
