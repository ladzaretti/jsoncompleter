package rejson_test

import (
	"fmt"
	"testing"

	"github.com/ladzaretti/rejson"
)

func TestReconstruct(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `   `,
			want:  `   `,
		},
		{
			input: `  a `,
			want:  `  a `,
		},
		{
			input: `  n`,
			want:  `  null`,
		},
		{
			input: `   fal  `,
			want:  `   false  `,
		},
		{
			input: `   1  `,
			want:  `   1  `,
		},
		{
			input: `{`,
			want:  `{}`,
		},
		{
			input: `{{{{[  `,
			want:  `{{{{[]}}}}  `,
		},
		{
			input: `[`,
			want:  `[]`,
		},
		{
			input: `{"nested": { "a": 1 }`,
			want:  `{"nested": { "a": 1 }}`,
		},
		{
			input: `{"key": "value`,
			want:  `{"key": "value"}`,
		},
		{
			input: `{"key": "esc\"ap\\ed`,
			want:  `{"key": "esc\"ap\\ed"}`,
		},
		{
			input: `{"key": "[[}{[[]]`,
			want:  `{"key": "[[}{[[]]"}`,
		},
		{
			input: `{"key": 1`,
			want:  `{"key": 1}`,
		},
		{
			input: `{"key": `,
			want:  `{"key": null} `,
		},
		{
			input: `{"key": f`,
			want:  `{"key": false}`,
		},
		{
			input: `{"key": tr`,
			want:  `{"key": true}`,
		},
		{
			input: `{"key": true}`,
			want:  `{"key": true}`,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Reconstruct(%s)", tt.input), func(t *testing.T) {
			if got := rejson.Reconstruct(tt.input); got != tt.want {
				t.Errorf("Reconstruct(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
