package jsoncompleter_test

import (
	"encoding/json"
	"fmt"

	"github.com/ladzaretti/jsoncompleter"
)

func Example_complete() {
	truncated := `{"key":"value","array":[1,2,3,4],"nested":{"key`

	completed := jsoncompleter.Complete(truncated)

	var anything any
	if err := json.Unmarshal([]byte(completed), &anything); err != nil {
		fmt.Printf("Error unmarshalling completed JSON: %v", err)
	}

	fmt.Printf("Completed json: %q", completed)
	// output: Completed json: "{\"key\":\"value\",\"array\":[1,2,3,4],\"nested\":{\"key\": \"\"}}"
}
