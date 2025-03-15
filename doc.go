// Package jsoncompleter provides utilities to complete truncated JSON strings.
//
// It provides the [Completer] struct, which is used to restore a valid JSON
// string from a truncated one.
//
// Use the [Complete] function for a quick way to complete a truncated JSON
// string, or create a new [Completer] with the [New] function when reusability
// is needed.
package jsoncompleter
