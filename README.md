# bodyguard

`bodyguard` is a Go library for JSON API response assertions in tests. It provides a declarative way to validate JSON structures, including support for exact matches, type checks, and complex matchers for strings, numbers, objects, and arrays.

## Installation

```bash
go get github.com/dedalusj/bodyguard
```

## Usage

`bodyguard` allows you to define the expected structure of your JSON response and then validate it against the actual response body.

### Basic Example

```go
package my_test

import (
	"testing"
	
	"github.com/dedalusj/bodyguard"
)

func TestExample_UserProfile(t *testing.T) {
	jsonPayload := `{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"username": "jdoe",
		"email": "jdoe@example.com",
		"age": 30,
		"active": true,
		"created_at": "2023-10-27T10:00:00Z",
		"address": {
			"street": "123 Main St",
			"city": "Anytown",
			"zip": "12345"
		},
		"tags": ["golang", "testing", "api"]
	}`

	// Use bodyguard.Assert to validate the structure
	bodyguard.Assert(t, bodyguard.Object(map[string]any{
		"id":         bodyguard.UUID(),
		"username":   "jdoe", // Exact match
		"email":      bodyguard.String(),
		"age":        bodyguard.NumberGreater(18),
		"active":     bodyguard.Bool(),
		"created_at": bodyguard.Timestamp(),
		"address": bodyguard.Object(map[string]any{
			"street": bodyguard.String(),
			"city":   "Anytown",
			"zip":    bodyguard.String(),
		}),
		"tags": bodyguard.Array("golang", bodyguard.String(), "api"),
	}), jsonPayload)
}
```

### Strict Validation

By default, `bodyguard.Object` allows extra fields in the JSON that are not defined in the matcher. If you want to ensure that *only* the specified fields are present, use `bodyguard.StrictObject`.

```go
func TestExample_StrictValidation(t *testing.T) {
	jsonPayload := `{
		"id": 1,
		"name": "Strict Item"
	}`

	bodyguard.Assert(t, bodyguard.StrictObject(map[string]any{
		"id":   1,
		"name": "Strict Item",
	}), jsonPayload)
}
```

### Unordered Arrays

If the order of elements in an array doesn't matter, use `bodyguard.UnorderedArray`.

```go
func TestExample_UnorderedList(t *testing.T) {
	jsonPayload := `["apple", "banana", "cherry"]`

	bodyguard.Assert(t, bodyguard.UnorderedArray(
		"cherry",
		"apple",
		"banana",
	), jsonPayload)
}
```

## Available Matchers

### Types
- `Null()`: Matches `null`.
- `Bool()`: Matches any boolean value.
- `String()`: Matches any string value.
- `Number()`: Matches any number value.
- `Object(map[string]any)`: Matches a JSON object.
- `StrictObject(map[string]any)`: Matches a JSON object exactly (no extra fields).
- `Array(...interface{})`: Matches a JSON array with elements in order.
- `UnorderedArray(...interface{})`: Matches a JSON array with elements in any order.

### String Matchers
- `UUID()`: Matches a string in UUID format.
- `Email()`: Matches a string in email format.
- `Regexp(pattern)`: Matches a string against a regular expression.
- `StringLength(min, max)`: Matches a string with length within the range.
- `URL()`: Matches a string in URL format.
- `OneOf(...options)`: Matches if the string is one of the options.
- `Timestamp()`: Matches a string in RFC3339 format.
- `Date()`: Matches a string in "2006-01-02" format.
- `StringWithFormat(func(string) error)`: Custom string format validator.

### Number Matchers
- `Number()`: Matches any number value.
- `Integer()`: Matches an integer value.
- `Positive()`: Matches a positive number.
- `Negative()`: Matches a negative number.
- `NumberWithinDelta(expected, delta)`: Matches a number within a delta of the expected value.
- `NumberWithinRange(min, max)`: Matches a number within the specified range (inclusive).
- `NumberGreater(min)`: Matches a number greater than the specified minimum.
- `NumberSmaller(max)`: Matches a number smaller than the specified maximum.

### Time Matchers
- `TimeWithinDuration(expected, delta)`: Matches a timestamp string within a duration of the expected time.
- `TimeWithinRange(startTime, endTime)`: Matches a timestamp string within the specified time range.
- `TimeBefore(before)`: Matches a timestamp string before the specified time.
- `TimeAfter(after)`: Matches a timestamp string after the specified time.

