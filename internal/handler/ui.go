package handler

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/a-h/templ"
)

// If returns value if condition is true, otherwise an empty value of type T.
func If[T comparable](condition bool, value T) T {
	var empty T
	if condition {
		return value
	}
	return empty
}

// IfElse returns trueValue if condition is true, otherwise falseValue.
func IfElse[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// MergeAttributes combines multiple templ.Attributes into one.
func MergeAttributes(attrs ...templ.Attributes) templ.Attributes {
	merged := templ.Attributes{}
	for _, attr := range attrs {
		for k, v := range attr {
			merged[k] = v
		}
	}
	return merged
}

// RandomID generates a random ID string for use in templ components.
func RandomID() string {
	return fmt.Sprintf("id-%s", rand.Text())
}

// scriptVersion is a timestamp generated at app start for cache busting.
var scriptVersion = fmt.Sprintf("%d", time.Now().Unix())

// ScriptURL generates a cache-busted URL for a static script path.
var ScriptURL = func(path string) string {
	return path + "?v=" + scriptVersion
}
