package grpc_api

import (
	"reflect"
	"testing"
)

func TestTriggerJobStateSortOrder(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "names", in: "error,idle,done,active", want: []string{"3", "0", "2", "1"}},
		{name: "raw codes kept", in: "2,0", want: []string{"2", "0"}},
		{name: "unknown kept", in: "idle,unknown", want: []string{"0", "unknown"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := triggerJobStateSortOrder(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("triggerJobStateSortOrder(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
