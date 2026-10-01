package grpc_api

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/webitel/engine/model"
)

func TestTransformAgent_ExtraChatCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count uint32
		want  string
	}{
		{
			name:  "zero is emitted",
			count: 0,
			want:  `"extra_chat_count":0`,
		},
		{
			name:  "non-zero is emitted",
			count: 3,
			want:  `"extra_chat_count":3`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res := transformAgent(&model.Agent{ExtraChatCount: tc.count})

			data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(res)
			if err != nil {
				t.Fatalf("marshal agent: %v", err)
			}

			if !strings.Contains(strings.ReplaceAll(string(data), " ", ""), tc.want) {
				t.Errorf("got %s, want it to contain %s", data, tc.want)
			}
		})
	}
}
