package grpc_api

import (
	"reflect"
	"testing"

	"github.com/webitel/engine/model"
)

var csvBOM = "\xEF\xBB\xBF"

func TestGenerateMembersCSVChunk(t *testing.T) {
	headers := []string{"name", "communications"}

	tests := []struct {
		name      string
		rows      [][]string
		page      int
		separator string
		want      string
	}{
		{
			name:      "default separator quotes values with comma",
			rows:      [][]string{{"john", "100, 200"}},
			page:      1,
			separator: "",
			want:      csvBOM + "name,communications\njohn,\"100, 200\"\n",
		},
		{
			name:      "single char separator quotes values containing it",
			rows:      [][]string{{"john", "100; 200"}, {"jane", "300"}},
			page:      1,
			separator: ";",
			want:      csvBOM + "name;communications\njohn;\"100; 200\"\njane;300\n",
		},
		{
			name:      "single char separator does not quote values with comma",
			rows:      [][]string{{"john", "100, 200"}},
			page:      1,
			separator: ";",
			want:      csvBOM + "name;communications\njohn;100, 200\n",
		},
		{
			name:      "tab separator",
			rows:      [][]string{{"john", "100\t200"}},
			page:      1,
			separator: "\t",
			want:      csvBOM + "name\tcommunications\njohn\t\"100\t200\"\n",
		},
		{
			name:      "multi-byte single rune separator",
			rows:      [][]string{{"john", "a§b"}},
			page:      1,
			separator: "§",
			want:      csvBOM + "name§communications\njohn§\"a§b\"\n",
		},
		{
			name:      "multi char separator keeps legacy join without quoting",
			rows:      [][]string{{"john", "100;;200"}},
			page:      1,
			separator: ";;",
			want:      csvBOM + "name;;communications\njohn;;100;;200\n",
		},
		{
			name:      "unsupported quote separator falls back to legacy join",
			rows:      [][]string{{"john", "100"}},
			page:      1,
			separator: "\"",
			want:      csvBOM + "name\"communications\njohn\"100\n",
		},
		{
			name:      "next page has no BOM and no headers",
			rows:      [][]string{{"jane", "300; 400"}},
			page:      2,
			separator: ";",
			want:      "jane;\"300; 400\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateMembersCSVChunk(headers, tt.rows, tt.page, tt.separator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMemberExportFieldValueCommunications(t *testing.T) {
	m := &model.Member{Communications: []*model.MemberCommunication{
		{Destination: "380501234567", Type: model.Lookup{Name: "Mobile"}, Priority: 0},
		{Destination: "380507654321", Priority: 1},
		{Destination: "380509999999", Type: model.Lookup{Name: "Work"}, Priority: 2},
	}}

	tests := map[string]string{
		"communications":           "380501234567, 380507654321, 380509999999",
		"communication_types":      "Mobile, , Work",
		"communication_priorities": "0, 1, 2",
	}
	for field, want := range tests {
		if got := memberExportFieldValue(m, field, nil); got != want {
			t.Errorf("%s: got %q, want %q", field, got, want)
		}
	}

	empty := &model.Member{}
	for field := range tests {
		if got := memberExportFieldValue(empty, field, nil); got != "" {
			t.Errorf("%s without communications: got %q, want empty", field, got)
		}
	}
}

func TestExportSearchFields(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty keeps defaults", nil, nil},
		{"no communications fields", []string{"id", "name"}, []string{"id", "name"}},
		{"types only", []string{"name", "communication_types"}, []string{"name", "communications"}},
		{"all three are deduplicated", []string{"communications", "communication_types", "communication_priorities", "id"},
			[]string{"communications", "id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exportSearchFields(tt.in)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandExportCommunicationFields(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"without communications", []string{"id", "name"}, []string{"id", "name"}},
		{"communications is expanded in place", []string{"id", "communications", "name"},
			[]string{"id", "communications", "communication_types", "communication_priorities", "name"}},
		{"explicit columns are not duplicated", []string{"communication_priorities", "communications"},
			[]string{"communication_priorities", "communications", "communication_types"}},
		{"all explicit", []string{"communications", "communication_types", "communication_priorities"},
			[]string{"communications", "communication_types", "communication_priorities"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandExportCommunicationFields(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
