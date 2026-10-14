package grpc_api

import (
	"reflect"
	"testing"
	"time"

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
		nil,
		{Destination: "380509999999", Type: model.Lookup{Name: "Work"}, Priority: 2},
	}}

	tests := map[string]string{
		"destination_1": "380501234567",
		"type_1":        "Mobile",
		"priority_1":    "0",
		"destination_2": "380507654321",
		"type_2":        "",
		"priority_2":    "1",
		"destination_3": "",
		"type_3":        "",
		"priority_3":    "",
		"destination_4": "380509999999",
		"type_4":        "Work",
		"priority_4":    "2",
		"destination_5": "",
		"priority_5":    "",
		"destination_0": "",
		"destination_x": "",
		"type_":         "",
		"unknown":       "",
	}
	for field, want := range tests {
		if got := memberExportFieldValue(m, field, nil); got != want {
			t.Errorf("%s: got %q, want %q", field, got, want)
		}
	}

	if got := memberExportFieldValue(&model.Member{}, "destination_1", nil); got != "" {
		t.Errorf("member without communications: got %q, want empty", got)
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
		{"destination alias", []string{"name", "destination"}, []string{"name", "communications"}},
		{"expanded columns collapse to one stored field",
			[]string{"id", "destination_1", "type_1", "priority_1", "destination_2", "name"},
			[]string{"id", "communications", "name"}},
		{"communications and its columns are deduplicated", []string{"communications", "type_1", "destination"},
			[]string{"communications"}},
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
		name  string
		in    []string
		count int
		want  []string
	}{
		{"without communications", []string{"id", "name"}, 3, []string{"id", "name"}},
		{"one communication", []string{"id", "communications", "name"}, 1,
			[]string{"id", "destination_1", "type_1", "priority_1", "name"}},
		{"several communications keep groups together", []string{"communications", "name"}, 2,
			[]string{"destination_1", "type_1", "priority_1", "destination_2", "type_2", "priority_2", "name"}},
		{"no communications in data still gives the first group", []string{"communications"}, 0,
			[]string{"destination_1", "type_1", "priority_1"}},
		{"destination alias is expanded once", []string{"destination", "communications"}, 1,
			[]string{"destination_1", "type_1", "priority_1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandExportCommunicationFields(tt.in, tt.count); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembersToExportRowsWithCommunicationColumns(t *testing.T) {
	members := []*model.Member{
		{Name: "john", Communications: []*model.MemberCommunication{
			{Destination: "100", Type: model.Lookup{Name: "Mobile"}, Priority: 0},
			{Destination: "200", Type: model.Lookup{Name: "Home"}, Priority: 1},
		}},
		{Name: "jane", Communications: []*model.MemberCommunication{
			{Destination: "300", Type: model.Lookup{Name: "Work"}, Priority: 5},
		}},
	}

	fields := expandExportCommunicationFields([]string{"name", "communications"}, 2)
	got := membersToExportRows(members, fields, time.UTC)
	want := [][]string{
		{"john", "100", "Mobile", "0", "200", "Home", "1"},
		{"jane", "300", "Work", "5", "", "", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
