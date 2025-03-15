//
// This is free and unencumbered software released into the public domain.
//
// Anyone is free to copy, modify, publish, use, compile, sell, or
// distribute this software, either in source code form or as a compiled
// binary, for any purpose, commercial or non-commercial, and by any
// means.
//
// In jurisdictions that recognize copyright laws, the author or authors
// of this software dedicate any and all copyright interest in the
// software to the public domain. We make this dedication for the benefit
// of the public at large and to the detriment of our heirs and
// successors. We intend this dedication to be an overt act of
// relinquishment in perpetuity of all present and future rights to this
// software under copyright law.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// For more information, please refer to <https://unlicense.org/>

package jsoncompleter

import (
	"strings"
	"unicode"
)

const JsonTruncatedMarker = `"__TRUNCATED__"`

// The JSON specification.
//
// Source: https://www.json.org/json-en.html

// Complete creates a new [Completer] and calls its [Completer.Complete] method
// with the provided input. It assumes the input is a truncated JSON string.
func Complete(truncated string) string {
	c := &Completer{}
	return c.Complete(truncated)
}

func New() *Completer {
	return &Completer{}
}

type Completer struct {
	openBrackets stack[rune]

	expectingKey, expectingColon, expectingValue bool
	insideQuotes                                 bool

	expectingEscape, expectingHex bool
	hexDigitsSeen                 int
}

// Complete processes a truncated JSON string and completes it into a valid JSON.
// It assumes that the input is a valid but truncated JSON string.
//
// Garbage in, garbage out:
// If the input is not a truncated JSON, the output is unpredictable.
func (c *Completer) Complete(input string) string {
	if len(input) == 0 {
		return input
	}

	l, r := leadingSpacesEnd(input), trailingSpacesStart(input)
	if r < l {
		return input // input is all spaces
	}

	leadingSpaces, trimmed, trailingSpaces := input[:l], input[l:r+1], input[r+1:]

	j := trimmed[0]
	if j != '{' && j != '[' {
		// not a json or an array value, so it is either not an invalid json
		// string or a truncated json literal (i.e., true, false, or null).
		if literal, ok := completeLiteral(trimmed); ok {
			return leadingSpaces + literal + trailingSpaces
		}
		return trimmed
	}

	defer c.reset()

	c.analyze(trimmed)

	return leadingSpaces + c.outputFrom(trimmed) + trailingSpaces
}

func leadingSpacesEnd(input string) int {
	i := 0
	for i < len(input) && unicode.IsSpace(rune(input[i])) {
		i++
	}
	return i
}

func trailingSpacesStart(input string) int {
	i := len(input) - 1
	for i >= 0 && unicode.IsSpace(rune(input[i])) {
		i--
	}
	return i
}

func (c *Completer) reset() {
	*c = Completer{}
}

func (c *Completer) analyze(input string) {
	for _, ch := range input {
		if ch == '"' {
			c.analyzeQuote()
			continue
		}

		if c.insideQuotes {
			c.analyzeString(ch)
			continue
		}

		c.analyzeStructural(ch)
	}
}

func (c *Completer) analyzeQuote() {
	if c.expectingEscape {
		c.expectingEscape = false
		return
	}

	c.insideQuotes = !c.insideQuotes

	if !c.insideQuotes {
		// we just closed a string value,
		// if it is a objects key, we now expect a colon.
		if c.insideObject() && c.expectingKey {
			c.expectingColon = true
		}
		c.expectingKey = false
		c.expectingValue = false
	}
}

func (c *Completer) analyzeString(ch rune) {
	switch ch {
	case '\\':
		c.expectingEscape = !c.expectingEscape
	case 'u':
		if c.expectingEscape {
			c.expectingEscape = false
			c.expectingHex = true
		}
	default:
		c.expectingEscape = false
		if c.expectingHex {
			c.handleHex()
		}
	}
}

func (c *Completer) handleHex() {
	c.hexDigitsSeen++
	if c.hexDigitsSeen == 4 {
		c.hexDigitsSeen = 0
		c.expectingHex = false
	}
}

func (c *Completer) analyzeStructural(ch rune) {
	switch ch {
	case '{':
		c.openBrackets.push('{')
		c.expectingKey = true
	case '[':
		c.openBrackets.push('[')
	case ':':
		c.expectingColon = false
	case ',':
		if c.insideObject() {
			c.expectingKey = true
		}
	case '}':
		if c.insideObject() {
			c.openBrackets.pop()
		}
	case ']':
		if c.insideArray() {
			c.openBrackets.pop()
		}
	}
}

func (c *Completer) insideObject() bool {
	if top, ok := c.openBrackets.peek(); ok && top == '{' {
		return true
	}

	return false
}

func (c *Completer) insideArray() bool {
	if top, ok := c.openBrackets.peek(); ok && top == '[' {
		return true
	}

	return false
}

func (c *Completer) outputFrom(input string) (output string) {
	output = input
	defer func() {
		output += c.balanceBrackets()
	}()

	lastCh := output[len(output)-1]
	if c.insideQuotes {
		output += c.completeString(lastCh)
		return
	}

	// remove trailing comma
	if lastCh == ',' {
		output = output[:len(output)-1]
		return
	}

	if val := c.completeMissingValue(lastCh); len(val) > 0 {
		output += val
		return
	}

	if literal, ok := completeLiteral(lastWord(output)); ok {
		output += literal
		return
	}

	if num := completeNumber(lastCh); len(num) > 0 {
		output += num
		return
	}

	return
}

func (c *Completer) completeString(last byte) (missing string) {
	var sb strings.Builder

	if c.expectingEscape {
		sb.WriteString("\\")
	}

	if c.expectingHex || c.hexDigitsSeen > 0 {
		sb.WriteString(strings.Repeat("0", 4-c.hexDigitsSeen))
	}

	switch {
	case c.expectingKey && last == '"':
		sb.WriteString(`key": null`)
	case c.expectingKey:
		sb.WriteString(`": null`)
	default:
		sb.WriteString(`"`)
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

func (c *Completer) completeMissingValue(last byte) string {
	if c.expectingColon {
		return ": null"
	}

	if last == ':' {
		return " null"
	}

	return ""
}

func completeNumber(last byte) string {
	switch last {
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

func (c *Completer) balanceBrackets() string {
	var sb strings.Builder

	for !c.openBrackets.empty() {
		bracket, _ := c.openBrackets.pop()
		switch bracket {
		case '{':
			sb.WriteRune('}')
		case '[':
			sb.WriteRune(']')
		}
	}

	return sb.String()
}
