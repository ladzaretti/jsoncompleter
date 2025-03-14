package truncatedjson_test

import (
	"encoding/json"
	"testing"

	"github.com/ladzaretti/truncatedjson"
)

func TestComplete(t *testing.T) {
	// jsonString := "[-123.0e+3]"
	jsonString := `
		{
		  "string": "text",
		  "hexString": "\u0000",
		  "number": 123,
		  "number": -123.0e+5,
		  "boolean_true": true,
		  "boolean_false": false,
		  "null_value": null,
		  "object": {
		    "nested_string": "ne{st[[]ed",
		    "nested_number": 42,
		    "nested_boolean_true": true,
		    "nested_boolean_false": false,
		    "nested_null": null,
		    "nested_array": ["it\\\\em1", "\u0000", 2, true, false, null, { "nested_key": "nested_value" }, [1, 2, 3, "\n"]]
		  },
		  "array": ["item1", 2, true, false, null, { "nested_key": "nested_value" }, [1, 2, 3]]
		}
	`

	for i := len(jsonString); i > 0; i-- {
		truncated := jsonString[:i]
		got := truncatedjson.Complete(truncated)

		if len(got) == len(truncated) {
			continue
		}

		var j any
		if err := json.Unmarshal([]byte(got), &j); err != nil {
			t.Errorf("Reconstruct(%q) produced invalid JSON: %v (output: %q)", truncated, err, got)
		}
	}
}
