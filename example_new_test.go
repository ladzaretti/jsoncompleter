package jsoncompleter_test

import (
	"encoding/json"
	"fmt"

	"github.com/ladzaretti/jsoncompleter"
)

func Example_new() {
	truncated := `{"key":"value","array":[1,2,3,`

	completer := jsoncompleter.New(
		jsoncompleter.WithMarkTruncation(true),
	)

	completed := completer.Complete(truncated)

	var anything any
	if err := json.Unmarshal([]byte(completed), &anything); err != nil {
		fmt.Printf("Error unmarshalling completed JSON: %v", err)
	}

	fmt.Printf("Completed json: %q", completed)
	// output: Completed json: "{\"key\":\"value\",\"array\":[1,2,3, \"__TRUNCATION_MARKER__\"]}"
}
