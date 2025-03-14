package truncatedjson

import (
	"strings"
	"unicode"
)

// The JSON specification.
//
// Source: https://www.json.org/json-en.html

// Complete attempts to reconstruct a truncated JSON string.
// It is assumed that the input is a valid JSON string that was truncated.
// No guarantees are made for non-JSON input.
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
	openBrackets *Stack[rune]

	expectingKey, expectingColon, expectingValue bool
	insideQuotes                                 bool

	expectingEscape, expectingHex bool
	hexDigitsSeen                 int
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

	tj.analyze(input)

	return tj.completeOutput(input)
}

func (tj *truncatedJSON) analyze(input string) {
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
}

func (tj *truncatedJSON) completeOutput(input string) (output string) {
	output = input
	defer func() {
		output += tj.balanceBrackets()
	}()

	lastCh := output[len(output)-1]
	if tj.insideQuotes {
		output += tj.completeString(lastCh)
		return
	}

	// remove trailing comma
	if lastCh == ',' {
		output = output[:len(output)-1]
		return
	}

	if s := tj.completeMissingValue(lastCh); len(s) > 0 {
		output += s
		return
	}

	if s := completeLiteralOrNumber(lastWord(output)); len(s) > 0 {
		output += s
		return
	}

	return
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

func (tj *truncatedJSON) completeMissingValue(last byte) string {
	if tj.expectingColon {
		return ": null"
	}

	if last == ':' {
		return " null"
	}

	return ""
}

func (tj *truncatedJSON) handleQuote() {
	if tj.expectingEscape {
		tj.expectingEscape = false
		return
	}

	tj.insideQuotes = !tj.insideQuotes

	if !tj.insideQuotes {
		// we just closed a string value,
		// if it is a objects key, we now expect a colon.
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
		tj.expectingEscape = !tj.expectingEscape
	case 'u':
		if tj.expectingEscape {
			tj.expectingEscape = false
			tj.expectingHex = true
		}
	default:
		tj.expectingEscape = false
		if tj.expectingHex {
			tj.handleHex()
		}
	}
}

func (tj *truncatedJSON) handleHex() {
	tj.hexDigitsSeen++
	if tj.hexDigitsSeen == 4 {
		tj.hexDigitsSeen = 0
		tj.expectingHex = false
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

	if tj.expectingEscape {
		sb.WriteString("\\")
	}

	if tj.expectingHex || tj.hexDigitsSeen > 0 {
		sb.WriteString(strings.Repeat("0", 4-tj.hexDigitsSeen))
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
