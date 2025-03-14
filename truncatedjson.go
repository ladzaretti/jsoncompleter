package truncatedjson

import (
	"strings"
	"unicode"
)

// Complete attempts to complete a truncated json string by completing
// incomplete literals and balancing brackets.
// The json spec https://www.json.org/json-en.html
func Complete(input string) string {
	if len(input) == 0 {
		return input
	}

	l := 0
	for l < len(input) && unicode.IsSpace(rune(input[l])) {
		l++
	}
	leadingSpaces := input[:l]

	r := len(input) - 1
	for r >= 0 && unicode.IsSpace(rune(input[r])) {
		r--
	}
	trailingSpaces := input[r+1:]

	if r < l {
		return input // all spaces
	}

	trimmed := input[l : r+1]

	tj := newTruncatedJSON()

	return leadingSpaces + tj.complete(trimmed) + trailingSpaces
}

type truncatedJSON struct {
	openBrackets                                 *Stack[rune]
	expectingKey, expectingColon, expectingValue bool
	insideQuotes                                 bool
	escape                                       bool
	hex                                          bool
	hexDigits                                    int
}

func newTruncatedJSON() *truncatedJSON {
	return &truncatedJSON{openBrackets: NewStack[rune]()}
}

func (tj *truncatedJSON) complete(input string) string {
	if len(input) == 0 {
		return input
	}

	j := input[0]
	if j != '{' && j != '[' {
		// not a json or an array value, so it is either not an invalid json
		// string or a truncated json literal (i.e., true, false, or null).
		if literal, ok := completeLiteral(input); ok {
			return literal
		}
		return input
	}

	for _, ch := range input {
		if ch == '"' {
			tj.handleQuote()
			continue
		}

		if tj.insideQuotes {
			tj.handleString(ch)
			continue
		}

		tj.handleStructural(ch)
	}

	output := input

	lastCh := output[len(output)-1]
	if tj.insideQuotes {
		output += tj.completeString(lastCh)
	}

	// remove trailing comma
	if lastCh == ',' {
		output = output[:len(output)-1]
	}

	output += tj.competeMissingValue(lastCh)

	output += completeLiteralOrNumber(lastWord(output))

	output += tj.balanceBrackets()

	return output
}

func (tj *truncatedJSON) insideObject() bool {
	if top, ok := tj.openBrackets.Peek(); ok && top == '{' {
		return true
	}

	return false
}

func (tj *truncatedJSON) insideArray() bool {
	if top, ok := tj.openBrackets.Peek(); ok && top == '[' {
		return true
	}

	return false
}

func (tj *truncatedJSON) competeMissingValue(last byte) string {
	if tj.expectingColon {
		return ": null"
	}

	if last == ':' {
		return " null"
	}

	return ""
}

func (tj *truncatedJSON) handleQuote() {
	if tj.escape {
		tj.escape = false
		return
	}

	tj.insideQuotes = !tj.insideQuotes

	if !tj.insideQuotes {
		// we just closed a string value,
		// if it is a objects key, we no expect a colon.
		if tj.insideObject() && tj.expectingKey {
			tj.expectingColon = true
		}
		tj.expectingKey = false
		tj.expectingValue = false
	}
}

func (tj *truncatedJSON) handleString(ch rune) {
	switch ch {
	case '\\':
		tj.escape = !tj.escape
	case 'u':
		if tj.escape {
			tj.escape = false
			tj.hex = true
		}
	default:
		tj.escape = false
		if tj.hex {
			tj.hexDigits++
			if tj.hexDigits == 4 {
				tj.hexDigits = 0
				tj.hex = false
			}
		}
	}
}

func (tj *truncatedJSON) handleStructural(ch rune) {
	switch ch {
	case '{':
		tj.openBrackets.Push('{')
		tj.expectingKey = true
	case '[':
		tj.openBrackets.Push('[')
	case ':':
		tj.expectingColon = false
	case ',':
		if tj.insideObject() {
			tj.expectingKey = true
		}
	case '}':
		if tj.insideObject() {
			tj.openBrackets.Pop()
		}
	case ']':
		if tj.insideArray() {
			tj.openBrackets.Pop()
		}
	}
}

func (tj *truncatedJSON) completeString(last byte) (missing string) {
	var sb strings.Builder

	if tj.escape {
		sb.WriteString("\\")
	}

	if tj.hex || tj.hexDigits > 0 {
		sb.WriteString(strings.Repeat("0", 4-tj.hexDigits))
	}

	switch {
	case tj.expectingKey && last == '"':
		sb.WriteString(`key": null`)
	case tj.expectingKey:
		sb.WriteString(`": null`)
	default:
		sb.WriteString(`"`)
	}
	return sb.String()
}

func (tj *truncatedJSON) balanceBrackets() string {
	var sb strings.Builder

	for !tj.openBrackets.Empty() {
		bracket, _ := tj.openBrackets.Pop()
		switch bracket {
		case '{':
			sb.WriteRune('}')
		case '[':
			sb.WriteRune(']')
		}
	}

	return sb.String()
}

var literals = map[string]string{
	"n":     "ull",
	"nu":    "ll",
	"nul":   "l",
	"null":  "",
	"t":     "rue",
	"tr":    "ue",
	"tru":   "e",
	"true":  "",
	"f":     "alse",
	"fa":    "lse",
	"fal":   "se",
	"fals":  "e",
	"false": "",
}

func completeLiteral(s string) (string, bool) {
	completed, ok := literals[s]
	return completed, ok
}

func completeLiteralOrNumber(word string) string {
	if len(word) == 0 {
		return ""
	}

	if completedLiteral, ok := completeLiteral(word); ok {
		return completedLiteral
	}

	switch word[len(word)-1] {
	case '-', '+', '.':
		return "0"
	case 'e', 'E':
		return "+0"
	default:
		return ""
	}
}

func lastWord(input string) string {
	i := len(input) - 1
	for i > 0 && !unicode.IsSpace(rune(input[i])) {
		i--
	}

	if i < len(input)-1 {
		return input[i+1:]
	}

	return ""
}
