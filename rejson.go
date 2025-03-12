package rejson

import (
	"unicode"
)

var literals = map[string]string{
	"n":    "ull",
	"nu":   "ll",
	"nul":  "l",
	"t":    "rue",
	"tr":   "ue",
	"tru":  "e",
	"f":    "alse",
	"fa":   "lse",
	"fal":  "se",
	"fals": "e",
}

func completeLiteral(s string) string {
	if len(s) > 5 { // max json literal length
		return ""
	}
	return literals[s]
}

// Reconstruct attempts to complete a truncated json string by completing
// incomplete literals and balancing brackets.
func Reconstruct(input string) string {
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

	return leadingSpaces + reconstruct(trimmed) + trailingSpaces
}

func reconstruct(input string) string {
	if len(input) == 0 {
		return input
	}

	output := input

	j := output[0]
	if j != '{' && j != '[' {
		// not a json object or array, so it is either not a json
		// string or a truncated json literal (i.e., true, false, or null).
		return output + completeLiteral(output)
	}

	openBrackets := NewStack[rune]()
	openQuotes := false
	expectingKey := false
	expectingColon := false

	for i, ch := range input {
		if openQuotes && ch != '"' {
			continue
		}

		switch ch {
		case '"':
			if i > 0 && input[i-1] == '\\' {
				break // ignore escaped quote
			}

			openQuotes = !openQuotes

			if !openQuotes {
				if insideObject(openBrackets) && expectingKey { // closing key quotes
					expectingColon = true
				}
				expectingKey = false
			}
		case ':':
			expectingColon = false
		case '{':
			openBrackets.Push('{')
			expectingKey = true
		case '}':
			if insideObject(openBrackets) {
				openBrackets.Pop()
			}
		case ',':
			if insideObject(openBrackets) {
				expectingKey = true
			}
		case '[':
			openBrackets.Push('[')
		case ']':
			if insideArray(openBrackets) {
				openBrackets.Pop()
			}
		default:
			if expectingKey && !openQuotes && !unicode.IsSpace(ch) {
				return input
			}
		}
	}

	last := output[len(output)-1]
	if openQuotes {
		switch {
		case expectingKey && last == '"':
			output += `key": null`
		case expectingKey:
			output += `": null`
		default:
			output += `"`
		}
	}

	last = output[len(output)-1]
	if last == ',' {
		output = output[:len(output)-1]
	}

	if expectingColon {
		output += ": null"
	}

	// append "null" if a json literal is expected but is missing
	if last == ':' {
		output += " null"
	}

	// complete a truncated json literal
	i := len(output) - 1
	for i > 0 && unicode.IsLower(rune(output[i])) {
		i--
	}

	// ensure there is a truncated literal to complete
	if i < len(output)-1 {
		literal := output[i+1:]
		output += completeLiteral(literal)
	}

	// balance brackets
	for !openBrackets.Empty() {
		bracket, _ := openBrackets.Pop()
		switch bracket {
		case '{':
			output += `}`
		case '[':
			output += `]`
		}
	}

	return output
}

func insideObject(openBrackets *Stack[rune]) bool {
	if top, ok := openBrackets.Peek(); ok && top == '{' {
		return true
	}

	return false
}

func insideArray(openBrackets *Stack[rune]) bool {
	if top, ok := openBrackets.Peek(); ok && top == '[' {
		return true
	}

	return false
}
