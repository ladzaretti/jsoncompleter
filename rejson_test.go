package rejson_test

import (
	"encoding/json"
	"testing"

	"github.com/ladzaretti/rejson"
)

func TestReconstruct(t *testing.T) {
	jsonString := `
		{
		  "string": "text",
		  "number": 123,
		  "boolean_true": true,
		  "boolean_false": false,
		  "null_value": null,
		  "object": {
		    "nested_string": "nested",
		    "nested_number": 42,
		    "nested_boolean_true": true,
		    "nested_boolean_false": false,
		    "nested_null": null,
        	    "nested_array": ["item1", 2, true, false, null, { "nested_key": "nested_value" }, [1, 2, 3]]
		  },
		  "array": ["item1", 2, true, false, null, { "nested_key": "nested_value" }, [1, 2, 3]]
		}
	`

	for i := len(jsonString); i > 0; i-- {
		truncated := jsonString[:i]
		got := rejson.Reconstruct(truncated)

		if len(got) == len(truncated) {
			continue
		}

		var j any
		if err := json.Unmarshal([]byte(got), &j); err != nil {
			t.Errorf("Reconstruct(%q) produced invalid JSON: %v (output: %q)", truncated, err, got)
		}
	}
}
