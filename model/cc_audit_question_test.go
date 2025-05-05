package model

import (
	"testing"
)

func TestQuestions_SumMin(t *testing.T) {
	tests := []struct {
		name     string
		qs       Questions
		required bool
		want     float32
	}{
		{
			name: "empty questions",
			qs:   Questions{},
			want: 0,
		},
		{
			name: "single required score question",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      1,
					Max:      5,
				},
			},
			required: true,
			want:     1,
		},
		{
			name: "single required options question",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Poor", Score: 1},
						{Name: "Good", Score: 3},
						{Name: "Excellent", Score: 5},
					},
				},
			},
			required: true,
			want:     1,
		},
		{
			name: "mixed questions - required only",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      1,
					Max:      5,
				},
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Poor", Score: 2},
						{Name: "Good", Score: 4},
					},
				},
				{
					Type:     QuestionTypeScore,
					Required: false,
					Min:      0,
					Max:      10,
				},
			},
			required: true,
			want:     3, // 1 + 2
		},
		{
			name: "mixed questions - non-required only",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: false,
					Min:      2,
					Max:      8,
				},
				{
					Type:     QuestionTypeOptions,
					Required: false,
					Options: []QuestionOption{
						{Name: "Low", Score: 1.5},
						{Name: "High", Score: 4.5},
					},
				},
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      0,
					Max:      10,
				},
			},
			required: false,
			want:     3.5, // 2 + 1.5
		},
		{
			name: "options with negative scores",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: -2},
						{Name: "Neutral", Score: 0},
						{Name: "Good", Score: 2},
					},
				},
			},
			required: true,
			want:     -2,
		},
		{
			name: "options with negative scores",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: -2},
						{Name: "Neutral", Score: 0},
						{Name: "Good", Score: 2},
					},
				},
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: -1},
						{Name: "Neutral", Score: 0},
						{Name: "Good", Score: 1},
					},
				},
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: 4},
						{Name: "Neutral", Score: 5},
						{Name: "Good", Score: 6},
					},
				},
			},
			required: true,
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qs.SumMin(tt.required); got != tt.want {
				t.Errorf("Questions.SumMin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuestions_SumMax(t *testing.T) {
	tests := []struct {
		name     string
		qs       Questions
		required bool
		want     float32
	}{
		{
			name: "empty questions",
			qs:   Questions{},
			want: 0,
		},
		{
			name: "single required score question",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      1,
					Max:      5,
				},
			},
			required: true,
			want:     5,
		},
		{
			name: "single required options question",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Poor", Score: 1},
						{Name: "Good", Score: 3},
						{Name: "Excellent", Score: 5},
					},
				},
			},
			required: true,
			want:     5,
		},
		{
			name: "mixed questions - required only",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      1,
					Max:      5,
				},
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Poor", Score: 2},
						{Name: "Good", Score: 4},
					},
				},
				{
					Type:     QuestionTypeScore,
					Required: false,
					Min:      0,
					Max:      10,
				},
			},
			required: true,
			want:     9, // 5 + 4
		},
		{
			name: "mixed questions - non-required only",
			qs: Questions{
				{
					Type:     QuestionTypeScore,
					Required: false,
					Min:      2,
					Max:      8,
				},
				{
					Type:     QuestionTypeOptions,
					Required: false,
					Options: []QuestionOption{
						{Name: "Low", Score: 1.5},
						{Name: "High", Score: 4.5},
					},
				},
				{
					Type:     QuestionTypeScore,
					Required: true,
					Min:      0,
					Max:      10,
				},
			},
			required: false,
			want:     12.5, // 8 + 4.5
		},
		{
			name: "options with negative scores",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: -2},
						{Name: "Neutral", Score: 0},
						{Name: "Good", Score: 2},
					},
				},
			},
			required: true,
			want:     2,
		},
		{
			name: "options with negative scores",
			qs: Questions{
				{
					Type:     QuestionTypeOptions,
					Required: true,
					Options: []QuestionOption{
						{Name: "Bad", Score: -2},
						{Name: "Neutral", Score: -1},
						{Name: "Good", Score: -3},
					},
				},
			},
			required: true,
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qs.SumMax(tt.required); got != tt.want {
				t.Errorf("Questions.SumMax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinScoreOption(t *testing.T) {
	tests := []struct {
		name    string
		options []QuestionOption
		want    float32
	}{
		{
			name:    "empty options",
			options: []QuestionOption{},
			want:    0,
		},
		{
			name: "single option",
			options: []QuestionOption{
				{Name: "Single", Score: 3.5},
			},
			want: 3.5,
		},
		{
			name: "multiple options positive scores",
			options: []QuestionOption{
				{Name: "Good", Score: 5},
				{Name: "Average", Score: 3},
				{Name: "Poor", Score: 1},
			},
			want: 1,
		},
		{
			name: "multiple options with negative scores",
			options: []QuestionOption{
				{Name: "Very Bad", Score: -5},
				{Name: "Bad", Score: -3},
				{Name: "Neutral", Score: 0},
				{Name: "Good", Score: 3},
			},
			want: -5,
		},
		{
			name: "all negative scores",
			options: []QuestionOption{
				{Name: "Terrible", Score: -5},
				{Name: "Very Bad", Score: -3},
				{Name: "Bad", Score: -1},
			},
			want: -5,
		},
		{
			name: "decimal scores",
			options: []QuestionOption{
				{Name: "High", Score: 4.5},
				{Name: "Medium", Score: 2.5},
				{Name: "Low", Score: 1.5},
			},
			want: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minScoreOption(tt.options); got != tt.want {
				t.Errorf("minScoreOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxScoreOption(t *testing.T) {
	tests := []struct {
		name    string
		options []QuestionOption
		want    float32
	}{
		{
			name:    "empty options",
			options: []QuestionOption{},
			want:    0,
		},
		{
			name: "single option",
			options: []QuestionOption{
				{Name: "Single", Score: 3.5},
			},
			want: 3.5,
		},
		{
			name: "multiple options positive scores",
			options: []QuestionOption{
				{Name: "Good", Score: 5},
				{Name: "Average", Score: 3},
				{Name: "Poor", Score: 1},
			},
			want: 5,
		},
		{
			name: "multiple options with negative scores",
			options: []QuestionOption{
				{Name: "Very Bad", Score: -5},
				{Name: "Bad", Score: -3},
				{Name: "Neutral", Score: 0},
				{Name: "Good", Score: 3},
			},
			want: 3,
		},
		{
			name: "all negative scores",
			options: []QuestionOption{
				{Name: "Terrible", Score: -5},
				{Name: "Very Bad", Score: -3},
				{Name: "Bad", Score: -1},
			},
			want: -1,
		},
		{
			name: "decimal scores",
			options: []QuestionOption{
				{Name: "High", Score: 4.5},
				{Name: "Medium", Score: 2.5},
				{Name: "Low", Score: 1.5},
			},
			want: 4.5,
		},
		{
			name: "decimal scores",
			options: []QuestionOption{
				{Name: "High", Score: 4.5},
				{Name: "Medium", Score: 2.5},
				{Name: "Low", Score: 1.5},
			},
			want: 4.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxScoreOption(tt.options); got != tt.want {
				t.Errorf("maxScoreOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
