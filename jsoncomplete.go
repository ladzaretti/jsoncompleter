package jsoncomplete

import (
	"unicode"
)

var literals = map[string]string{
	"n":    "null",
	"nu":   "null",
	"nul":  "null",
	"t":    "true",
	"tr":   "true",
	"tru":  "true",
	"f":    "false",
	"fa":   "false",
	"fal":  "false",
	"fals": "false",
}

func completeLiteral(s string) (string, bool) {
	if len(s) > 5 { // max json literal length
		return "", false
	}

	completed, ok := literals[s]

	return completed, ok
}

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

	return leadingSpaces + complete(trimmed) + trailingSpaces
}

func complete(input string) string {
	if len(input) == 0 {
		return input
	}

	j := input[0]
	if j != '{' && j != '[' {
		// not a json object or array, so it is either not a json
		// string or a truncated json literal (i.e., true, false, or null).
		if literal, ok := completeLiteral(input); ok {
			return literal
		}
		return input
	}

	var (
		output       = input
		openBrackets = NewStack[rune]()

		expectingKey, expectingColon = false, false
		escape, hex, hexDigits       = false, false, 0

		insideQuotes = false
		insideObject = func() bool {
			if top, ok := openBrackets.Peek(); ok && top == '{' {
				return true
			}

			return false
		}
		insideArray = func() bool {
			if top, ok := openBrackets.Peek(); ok && top == '[' {
				return true
			}

			return false
		}
	)

	for _, ch := range input {
		switch ch {
		case '"':
			if escape {
				escape = false
				break
			}

			insideQuotes = !insideQuotes

			if !insideQuotes {
				// we just closed a string value,
				// if it is a objects key, we no expect a colon.
				if insideObject() && expectingKey {
					expectingColon = true
				}
				expectingKey = false
			}
		case '\\':
			if insideQuotes {
				escape = !escape
			}
		case 'u':
			if escape {
				escape = false
				hex = true
			}
		case ':':
			expectingColon = false
		case '{':
			openBrackets.Push('{')
			expectingKey = true
		case '}':
			if insideObject() {
				openBrackets.Pop()
			}
		case ',':
			if insideObject() {
				expectingKey = true
			}
		case '[':
			openBrackets.Push('[')
		case ']':
			if insideArray() {
				openBrackets.Pop()
			}
		default:
			escape = false
			if hex {
				hexDigits++
				if hexDigits == 4 {
					hexDigits = 0
					hex = false
				}
			}
		}
	}

	if escape {
		output += "\\"
	}

	if hex || hexDigits > 0 {
		for i := hexDigits; i < 4; i++ {
			output += "0"
		}
	}

	last := output[len(output)-1]
	if insideQuotes {
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

	// complete a possibly truncated json literal
	i := len(output) - 1
	for i > 0 && unicode.IsLower(rune(output[i])) {
		i--
	}

	if i < len(output)-1 { // ensure there is a truncated maybeLiteral to complete
		maybeLiteral := output[i+1:]
		if completedLiteral, ok := completeLiteral(maybeLiteral); ok {
			output = output[:i]
			output += completedLiteral
		}
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
