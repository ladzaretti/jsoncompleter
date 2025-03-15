package jsoncompleter_test

import (
	"encoding/json"
	"fmt"

	"github.com/ladzaretti/jsoncompleter"
)

func Example_new() {
	truncated := `{"key":"value","array":[1,2,3,4],"nested":{"key1":"value1",`

	completer := jsoncompleter.New()

	completed := completer.Complete(truncated)

	var anything any
	if err := json.Unmarshal([]byte(completed), &anything); err != nil {
		fmt.Printf("Error unmarshalling completed JSON: %v", err)
	}

	fmt.Printf("Completed json: %q", completed)
	// output: Completed json: "{\"key\":\"value\",\"array\":[1,2,3,4],\"nested\":{\"key1\":\"value1\",\"key\":\"__TRUNCATED__\"}}"
}
