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

// The JSON specification.
//
// Source: https://www.json.org/json-en.html

// DefaultTruncationMarker is the default string for marking truncation
// in JSON when enabled.
const DefaultTruncationMarker = "__TRUNCATION_MARKER__"

// Complete creates a new [Completer] and calls its [Completer.Complete]
// receiver function with the provided input.
// It assumes the input is a truncated JSON string.
func Complete(truncated string) string {
	return New().Complete(truncated)
}

type Opt func(*Completer)

// WithTruncationMarker sets a custom string to mark where
// the JSON was truncated and fixed.
//
// The default value is [DefaultTruncationMarker].
func WithTruncationMarker(s string) Opt {
	return func(c *Completer) {
		c.config.truncationMarker = s
	}
}

// WithMarkTruncation enables marking the place where
// the JSON got truncated and fixed.
//
// Example:
//
//	Input (truncated JSON):
//		{ "key": "value", "array": [1,2,3
//
//	Output (after completion):
//		{ "key": "value", "array": [1,2,3,"__TRUNCATION_MARKER__"] }
func WithMarkTruncation(enabled bool) Opt {
	return func(c *Completer) {
		c.config.markTruncation = enabled
	}
}

func New(opts ...Opt) *Completer {
	c := &Completer{config: config{
		truncationMarker: DefaultTruncationMarker,
	}}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type config struct {
	truncationMarker string
	markTruncation   bool
}

type Completer struct {
	config config

	openBrackets stack[rune]

	expectingKey, expectingColon bool
	insideString                 bool

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
		// not a json object or array value, so it is either an invalid json
		// string or a truncated json primitive (i.e., true, false, or null).
		if literal, ok := completeBoolNull(trimmed); ok {
			return leadingSpaces + literal + trailingSpaces
		}

		return trimmed
	}

	defer c.reset()

	c.analyze(trimmed)

	return leadingSpaces + c.completeTruncated(trimmed) + trailingSpaces
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
	*c = Completer{config: c.config}
}

func (c *Completer) analyze(input string) {
	for _, ch := range input {
		if ch == '"' {
			c.analyzeStringBeginEnd()
			continue
		}

		if c.insideString {
			c.analyzeString(ch)
			continue
		}

		c.analyzeStructural(ch)
	}
}

func (c *Completer) analyzeStringBeginEnd() {
	if c.expectingEscape {
		c.expectingEscape = false
		return
	}

	c.insideString = !c.insideString

	if !c.insideString {
		// we just closed a string value,
		// if it is an object key, we now expect a colon.
		if c.insideObject() && c.expectingKey {
			c.expectingColon = true
		}

		c.expectingKey = false
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

//nolint:nakedret
func (c *Completer) completeTruncated(input string) (output string) {
	output = input
	defer func() {
		ch := output[len(output)-1]
		output += c.markTruncation(ch) + c.balanceBrackets()
	}()

	if c.insideString {
		output += c.completeString()
		if c.expectingKey {
			output += `: ""`
		}

		return
	}

	ch := output[len(output)-1]

	// remove trailing comma
	if ch == ',' {
		output = output[:len(output)-1]
		return
	}

	if value := c.completeValue(ch); len(value) > 0 {
		output += value
		return
	}

	if boolNull, ok := completeBoolNull(lastNonString(input)); ok {
		output += boolNull
		return
	}

	if number := completeNumber(ch); len(number) > 0 {
		output += number
		return
	}

	return
}

func (c *Completer) markTruncation(last byte) string {
	if !c.config.markTruncation {
		return ""
	}

	if c.insideArray() {
		if last == '[' {
			return `"` + c.config.truncationMarker + `"`
		}

		return `, "` + c.config.truncationMarker + `"`
	}

	if c.insideObject() {
		if last == '{' {
			return `"` + c.config.truncationMarker + `": ""`
		}

		return `, "` + c.config.truncationMarker + `": ""`
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

func (c *Completer) completeString() (missing string) {
	var sb strings.Builder

	if c.expectingEscape {
		sb.WriteString("\\")
	}

	if c.expectingHex || c.hexDigitsSeen > 0 {
		sb.WriteString(strings.Repeat("0", 4-c.hexDigitsSeen))
	}

	sb.WriteString(`"`)

	return sb.String()
}

func (c *Completer) completeValue(last byte) string {
	if c.expectingColon {
		return `: ""`
	}

	if last == ':' {
		return ` ""`
	}

	return ""
}

var boolNullSuffix = map[string]string{
	"n":     "ull",
	"nu":    "ll",
	"nul":   "l",
	"null":  "",
	"t":     "rue",
	"tr":    "ue",
	"tru":   "e",
	"true":  "",
	"f":     "alse", //nolint:misspell
	"fa":    "lse",
	"fal":   "se",
	"fals":  "e",
	"false": "",
}

func completeBoolNull(s string) (string, bool) {
	suffix, ok := boolNullSuffix[s]
	return suffix, ok
}

func completeNumber(last byte) string {
	switch last {
	case '-', '+', '.', 'e', 'E':
		return "0"
	default:
		return ""
	}
}

func lastNonString(input string) string {
	i := len(input) - 1

	for i > 0 {
		r := rune(input[i])

		if unicode.IsSpace(r) || isDelimiter(r) {
			break
		}

		i--
	}

	if i < len(input)-1 {
		return input[i+1:]
	}

	return ""
}

func isDelimiter(r rune) bool {
	switch r {
	case '{', '[', ',', ':', ']', '}':
		return true
	default:
		return false
	}
}
