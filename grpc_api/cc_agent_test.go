package grpc_api

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"github.com/webitel/engine/gen/engine"
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

func TestCreateAgentEmailDefaults(t *testing.T) {
	for _, tc := range []struct {
		input       string
		base, extra uint32
	}{
		{input: `{}`, base: 1},
		{input: `{"email_count":0}`},
		{input: `{"email_count":3}`, base: 3},
		{input: `{"email_count":0,"extra_email_count":2}`, extra: 2},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var in engine.CreateAgentRequest
			if err := protojson.Unmarshal([]byte(tc.input), &in); err != nil {
				t.Fatal(err)
			}
			if base := createAgentEmailCount(in.EmailCount); base != tc.base || in.GetExtraEmailCount() != tc.extra {
				t.Fatalf("Email limits = %d/%d, want %d/%d", base, in.GetExtraEmailCount(), tc.base, tc.extra)
			}
		})
	}
}

func TestUpdateAgentEmailLimitPresence(t *testing.T) {
	for _, tc := range []struct {
		input       string
		base, extra bool
	}{
		{input: `{}`},
		{input: `{"email_count":0}`, base: true},
		{input: `{"extra_email_count":0}`, extra: true},
		{input: `{"email_count":0,"extra_email_count":0}`, base: true, extra: true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var in engine.UpdateAgentRequest
			if err := protojson.Unmarshal([]byte(tc.input), &in); err != nil {
				t.Fatal(err)
			}
			if (in.EmailCount != nil) != tc.base || (in.ExtraEmailCount != nil) != tc.extra {
				t.Fatal("Email-limit presence lost")
			}
			if in.GetEmailCount() != 0 || in.GetExtraEmailCount() != 0 {
				t.Fatal("explicit zero changed")
			}
		})
	}
}

func TestTransformAgentEmailLimits(t *testing.T) {
	for _, count := range []uint32{0, 3} {
		res := transformAgent(&model.Agent{EmailCount: 2, ExtraEmailCount: count})
		data, err := protojson.Marshal(res)
		if err != nil {
			t.Fatal(err)
		}
		var decoded engine.Agent
		if err := protojson.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.GetEmailCount() != 2 || decoded.GetExtraEmailCount() != count {
			t.Fatalf("response Email limits = %d/%d, want 2/%d", decoded.GetEmailCount(), decoded.GetExtraEmailCount(), count)
		}
	}
}

func TestTransformAgentEmailPostponedCount(t *testing.T) {
	zero, five := uint32(0), uint32(5)
	for _, limit := range []*uint32{nil, &zero, &five} {
		data, err := protojson.Marshal(transformAgent(&model.Agent{EmailPostponedCount: limit}))
		if err != nil {
			t.Fatal(err)
		}
		var decoded engine.Agent
		if err := protojson.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if (decoded.EmailPostponedCount == nil) != (limit == nil) || (limit != nil && *decoded.EmailPostponedCount != *limit) {
			t.Fatalf("response postponed limit = %v, want %v (%s)", decoded.EmailPostponedCount, limit, data)
		}
	}
}

func TestAgentEmailPostponedCountWirePresence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		present bool
		count   uint64
	}{
		{name: "omitted"},
		{name: "explicit zero", present: true},
		{name: "nonzero", present: true, count: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			if tc.present {
				data = protowire.AppendVarint(protowire.AppendTag(nil, 17, protowire.VarintType), tc.count)
			}
			var create engine.CreateAgentRequest
			var patch engine.PatchAgentRequest
			for _, m := range []proto.Message{&create, &patch} {
				if err := proto.Unmarshal(data, m); err != nil {
					t.Fatal(err)
				}
			}
			for _, count := range []*uint32{create.EmailPostponedCount, patch.EmailPostponedCount} {
				if (count != nil) != tc.present || (count != nil && uint64(*count) != tc.count) {
					t.Fatal("postponed limit lost wire presence or value")
				}
			}
		})
	}
}

func TestAgentEmailCountWirePresence(t *testing.T) {
	for _, tc := range []struct {
		name    string
		present bool
		count   uint64
	}{
		{name: "omitted including legacy scalar zero"},
		{name: "encoded explicit zero", present: true},
		{name: "legacy nonzero", present: true, count: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, field := range []protowire.Number{16, 17} {
				var data []byte
				if tc.present {
					data = protowire.AppendVarint(protowire.AppendTag(nil, field, protowire.VarintType), tc.count)
				}
				var create engine.CreateAgentRequest
				var update engine.UpdateAgentRequest
				var message proto.Message = &create
				if field == 17 {
					message = &update
				}
				if err := proto.Unmarshal(data, message); err != nil {
					t.Fatal(err)
				}
				count := create.EmailCount
				if field == 17 {
					count = update.EmailCount
				}
				if (count != nil) != tc.present || (count != nil && uint64(*count) != tc.count) {
					t.Fatalf("field %d lost wire presence or value", field)
				}
			}
		})
	}
}
