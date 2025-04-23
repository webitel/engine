package model

import (
	"reflect"
	"testing"
)

func TestParseRegexp(t *testing.T) {
	type args struct {
		q string
	}
	tests := []struct {
		name      string
		args      args
		wantS     *string
		wantFound bool
	}{
		{
			name: "simple",
			args: args{
				q: "simple",
			},
			wantS:     &[]string{"%simple%"}[0],
			wantFound: false,
		},
		{
			name: "simple with suffix star",
			args: args{
				q: "simple*",
			},
			wantS:     &[]string{"%simple%"}[0],
			wantFound: false,
		},
		{
			name: "simple with prefix star",
			args: args{
				q: "*simple",
			},
			wantS:     &[]string{"%simple%"}[0],
			wantFound: false,
		},
		{
			name: "simple with double star",
			args: args{
				q: "*simple*",
			},
			wantS:     &[]string{"%simple%"}[0],
			wantFound: false,
		},
		{
			name: "simple with mixed regexp and star",
			args: args{
				q: "/simple*",
			},
			wantS:     &[]string{"%/simple%"}[0],
			wantFound: false,
		},

		{
			name: "simple with regexp",
			args: args{
				q: "/simple/",
			},
			wantS:     &[]string{"simple"}[0],
			wantFound: true,
		},
		{
			name: "simple with regexp and added star",
			args: args{
				q: "/simple/*",
			},
			wantS:     &[]string{"simple"}[0],
			wantFound: true,
		},
		{
			name: "simple with regexp",
			args: args{
				q: "/simple*/",
			},
			wantS:     &[]string{"simple*"}[0],
			wantFound: true,
		},
		{
			name: "simple with regexp",
			args: args{
				q: "/*simple*/",
			},
			wantS:     &[]string{"*simple*"}[0],
			wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, gotFound := ParseRegexp(tt.args.q)
			if !reflect.DeepEqual(*gotS, *tt.wantS) {
				t.Errorf("ParseRegexp() gotS = %v, want %v", *gotS, *tt.wantS)
			}
			if gotFound != tt.wantFound {
				t.Errorf("ParseRegexp() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}
