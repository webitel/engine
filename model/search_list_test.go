package model_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/webitel/engine/model"
)

func TestParseRegexp(t *testing.T) {
	type args struct {
		q string
	}

	t.Parallel()

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
			t.Parallel()

			gotS, gotFound := model.ParseRegexp(tt.args.q)
			if !reflect.DeepEqual(*gotS, *tt.wantS) {
				t.Errorf("ParseRegexp() gotS = %v, want %v", *gotS, *tt.wantS)
			}
			if gotFound != tt.wantFound {
				t.Errorf("ParseRegexp() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestNewFilterBetween(t *testing.T) {
	now := time.Now().UTC()
	nowMs := now.UnixMilli()
	weekAgoMs := now.AddDate(0, 0, -7).UnixMilli()

	fixedToMs := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC).UnixMilli()
	fixedFromMs := time.Date(2026, time.May, 10, 12, 0, 0, 0, time.UTC).UnixMilli()
	fixedToMinus7Ms := time.Date(2026, time.May, 13, 12, 0, 0, 0, time.UTC).UnixMilli()

	t.Parallel()

	tests := []struct {
		name     string
		from     int64
		to       int64
		wantFrom int64
		wantTo   int64
		deltaMs  int64
	}{
		{
			name:     "Both from and to are zero -> defaults to now-7d and now",
			from:     0,
			to:       0,
			wantFrom: weekAgoMs,
			wantTo:   nowMs,
			deltaMs:  50,
		},
		{
			name:     "To is set, From is zero -> From defaults to To - 7 days",
			from:     0,
			to:       fixedToMs,
			wantFrom: fixedToMinus7Ms,
			wantTo:   fixedToMs,
			deltaMs:  0,
		},
		{
			name:     "From is set, To is zero -> To defaults to now",
			from:     fixedFromMs,
			to:       0,
			wantFrom: fixedFromMs,
			wantTo:   nowMs,
			deltaMs:  50,
		},
		{
			name:     "Both from and to are set validly -> keeps values as is",
			from:     fixedFromMs,
			to:       fixedToMs,
			wantFrom: fixedFromMs,
			wantTo:   fixedToMs,
			deltaMs:  0,
		},
		{
			name:     "Invalid range: From > To -> corrects From to To - 7 days",
			from:     fixedToMs,
			to:       fixedFromMs,
			wantFrom: time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC).UnixMilli(),
			wantTo:   fixedFromMs,
			deltaMs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := model.NewFilterBetween(tt.from, tt.to)

			if got == nil {
				t.Fatalf("NewFilterBetween() returned nil")
			}

			if !withinDelta(got.From, tt.wantFrom, tt.deltaMs) {
				t.Errorf("NewFilterBetween().From = %v, want %v (±%d ms)", got.From, tt.wantFrom, tt.deltaMs)
			}

			if !withinDelta(got.To, tt.wantTo, tt.deltaMs) {
				t.Errorf("NewFilterBetween().To = %v, want %v (±%d ms)", got.To, tt.wantTo, tt.deltaMs)
			}
		})
	}
}

func withinDelta(got, want, delta int64) bool {
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	return diff <= delta
}
