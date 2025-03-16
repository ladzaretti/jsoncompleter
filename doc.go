// Package jsoncompleter provides the [Completer] struct with a single method
// to restore truncated JSON strings.
//
// Use the [Complete] function for a quick way to fix truncated JSON, or create a
// new [Completer] with the [New] function for reusable operations.
//
// Truncation is handled by adding missing elements or completing primitives.
// No data is deleted, except for one case: a trailing comma, which is removed.
//
// When truncation occurs within a number, 0 is added when needed—such as when
// truncated at a sign, a decimal point, or an exponent indicator.
//
// The null, true, and false primitives are completed to their full values
// (e.g., "n" -> "null").
//
// When adding missing strings or values, an empty string is used as a placeholder.
package jsoncompleter
