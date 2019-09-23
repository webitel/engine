package grpc_api

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/struct"
)

var (
	jsonpbCodec = struct {
		jsonpb.Unmarshaler
		jsonpb.Marshaler
	}{
		Unmarshaler: jsonpb.Unmarshaler{

			// Whether to allow messages to contain unknown fields, as opposed to
			// failing to unmarshal.
			AllowUnknownFields: false, // bool

			// A custom URL resolver to use when unmarshaling Any messages from JSON.
			// If unset, the default resolution strategy is to extract the
			// fully-qualified type name from the type URL and pass that to
			// proto.MessageType(string).
			AnyResolver: nil,
		},
		Marshaler: jsonpb.Marshaler{

			// Whether to render enum values as integers, as opposed to string values.
			EnumsAsInts: false, // bool

			// Whether to render fields with zero values.
			EmitDefaults: true, // bool

			// A string to indent each level by. The presence of this field will
			// also cause a space to appear between the field separator and
			// value, and for newlines to be appear between fields and array
			// elements.
			Indent: "", // string

			// Whether to use the original (.proto) name for fields.
			OrigName: true, // bool

			// A custom URL resolver to use when marshaling Any messages to JSON.
			// If unset, the default resolution strategy is to extract the
			// fully-qualified type name from the type URL and pass that to
			// proto.MessageType(string).
			AnyResolver: nil,
		},
	}
)

func UnmarshalJsonpb(data []byte) *structpb.Value {
	var pb structpb.Value
	rd := bytes.NewReader(data)
	err := jsonpbCodec.Unmarshal(rd, &pb)
	if err != nil {
		return nil
	}
	return &pb
}

func MarshalJsonpb(pb *structpb.Value) []byte {
	var buf bytes.Buffer
	err := jsonpbCodec.Marshal(&buf, pb)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
