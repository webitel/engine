package model_test

import (
	"reflect"
	"testing"

	"github.com/webitel/engine/model"
)

func comm(id int64, dest, typ string, prio int) *model.MemberCommunication {
	return &model.MemberCommunication{
		Id:          id,
		Destination: dest,
		Type:        model.Lookup{Name: typ},
		Priority:    prio,
	}
}

func TestMember_SortCommunications(t *testing.T) {
	t.Parallel()

	base := func() []*model.MemberCommunication {
		return []*model.MemberCommunication{
			comm(1, "Bob", "sip", 2),
			comm(2, "alice", "email", 5),
			comm(3, "Carol", "sip", 1),
		}
	}

	tests := []struct {
		name    string
		order   string
		input   []*model.MemberCommunication
		wantIDs []int64
	}{
		{"empty leaves order", "", base(), []int64{1, 2, 3}},
		{"unknown key leaves order", "weird", base(), []int64{1, 2, 3}},
		{"name asc case-insensitive", "name", base(), []int64{2, 1, 3}},
		{"name asc plus prefix", "+name", base(), []int64{2, 1, 3}},
		{"name asc space prefix", " name", base(), []int64{2, 1, 3}},
		{"name desc", "-name", base(), []int64{3, 1, 2}},
		{"destination alias", "destination", base(), []int64{2, 1, 3}},
		{"type asc stable", "type", base(), []int64{2, 1, 3}},
		{"priority asc", "priority", base(), []int64{3, 1, 2}},
		{"priority desc", "-priority", base(), []int64{2, 1, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &model.Member{Communications: tc.input}
			m.SortCommunications(tc.order)

			got := make([]int64, 0, len(m.Communications))
			for _, c := range m.Communications {
				got = append(got, c.Id)
			}
			if !reflect.DeepEqual(got, tc.wantIDs) {
				t.Fatalf("SortCommunications(%q) order = %v, want %v", tc.order, got, tc.wantIDs)
			}
		})
	}
}

func TestMember_SortCommunications_NilsLast(t *testing.T) {
	t.Parallel()

	m := &model.Member{Communications: []*model.MemberCommunication{
		comm(1, "b", "x", 1),
		nil,
		comm(2, "a", "y", 2),
	}}

	m.SortCommunications("name")

	if len(m.Communications) != 3 {
		t.Fatalf("len = %d, want 3", len(m.Communications))
	}
	if c := m.Communications[0]; c == nil || c.Id != 2 {
		t.Fatalf("first = %v, want id 2", c)
	}
	if c := m.Communications[1]; c == nil || c.Id != 1 {
		t.Fatalf("second = %v, want id 1", c)
	}
	if m.Communications[2] != nil {
		t.Fatalf("third = %v, want nil", m.Communications[2])
	}
}
