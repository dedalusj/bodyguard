package bodyguard

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"
	"time"
)

// Matcher is the core interface for all assertions.
// It matches a value found at a specific path in the JSON.
type Matcher interface {
	Match(path string, value interface{}) error
}

var (
	_ Matcher = MatcherFunc(nil)
)

// MatcherFunc is a helper for simple function-based matchers
type MatcherFunc func(path string, value interface{}) error

func (m MatcherFunc) Match(path string, value interface{}) error {
	return m(path, value)
}

// Assert checks that the given body (as a string or []byte) matches the expected structure.
// expected can be a Matcher, or a raw value (which will be strictly compared).
// It fails the test if there is a mismatch.
func Assert(t *testing.T, expected interface{}, body interface{}) {
	t.Helper()
	if err := isMatch(body, expected); err != nil {
		t.Error(err)
	}
}

func isMatch(body interface{}, expected interface{}) error {
	var actual interface{}
	var bodyBytes []byte

	switch b := body.(type) {
	case string:
		bodyBytes = []byte(b)
	case []byte:
		bodyBytes = b
	default:
		return fmt.Errorf("body must be string or []byte, got %T", body)
	}

	if err := json.Unmarshal(bodyBytes, &actual); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	return match(expected, "$", actual)
}

func match(expected interface{}, path string, actual interface{}) error {
	if m, ok := expected.(Matcher); ok {
		return m.Match(path, actual)
	}

	// Exact match handling for literals
	if reflect.DeepEqual(expected, actual) {
		return nil
	}

	// conversions for numbers which unmarshal as float64
	val := reflect.ValueOf(expected)
	matched := false
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f64, ok := actual.(float64); ok {
			if float64(val.Int()) == f64 {
				matched = true
			}
		}
	}

	if matched {
		return nil
	}

	return fmt.Errorf("at %s: expected %v (%T), got %v (%T)", path, expected, expected, actual, actual)
}

// Null asserts the value is null
func Null() Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		if value != nil {
			return fmt.Errorf("at %s: expected null, got %v", path, value)
		}
		return nil
	})
}

// Bool asserts the value is a boolean.
func Bool() Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("at %s: expected boolean, got %T", path, value)
		}
		return nil
	})
}

func stringValue(validators ...func(string) error) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("at %s: expected string, got %T", path, value)
		}
		for _, v := range validators {
			if err := v(s); err != nil {
				return fmt.Errorf("at %s: %w", path, err)
			}
		}
		return nil
	})
}

// String checks if the value is a string
func String() Matcher {
	return stringValue()
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// UUID checks if the value is a valid UUID string
func UUID() Matcher {
	return stringValue(func(s string) error {
		if !uuidRegex.MatchString(s) {
			return fmt.Errorf("expected UUID, got %q", s)
		}
		return nil
	})
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// Email checks if the value is a valid email string
func Email() Matcher {
	return stringValue(func(s string) error {
		if !emailRegex.MatchString(s) {
			return fmt.Errorf("expected email, got %q", s)
		}
		return nil
	})
}

// Regexp checks if the value matches the specified regular expression
func Regexp(pattern string) Matcher {
	re, err := regexp.Compile(pattern)
	return stringValue(func(s string) error {
		if err != nil {
			return fmt.Errorf("invalid regexp pattern %q: %w", pattern, err)
		}
		if !re.MatchString(s) {
			return fmt.Errorf("expected to match %q, got %q", pattern, s)
		}
		return nil
	})
}

// StringLength checks if the string length is within the specified range
func StringLength(min, max int) Matcher {
	return stringValue(func(s string) error {
		length := len(s)
		if length < min || length > max {
			return fmt.Errorf("expected string length between %d and %d, got %d", min, max, length)
		}
		return nil
	})
}

// URL checks if the value is a valid URL
func URL() Matcher {
	return stringValue(func(s string) error {
		if !regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`).MatchString(s) {
			return fmt.Errorf("expected valid URL, got %q", s)
		}
		return nil
	})
}

// OneOf checks if the value is one of the specified strings
func OneOf(options ...string) Matcher {
	return stringValue(func(s string) error {
		for _, opt := range options {
			if s == opt {
				return nil
			}
		}
		return fmt.Errorf("expected one of %v, got %q", options, s)
	})
}

// StringWithFormat checks if the value matches a custom string format
func StringWithFormat(formatCheck func(string) error) Matcher {
	return stringValue(formatCheck)
}

func timeValue(parser func(string) (time.Time, error), validators ...func(time.Time) error) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("at %s: expected time string, got %T", path, value)
		}

		parsed, err := parser(s)
		if err != nil {
			return fmt.Errorf("at %s: %w", path, err)
		}

		for _, v := range validators {
			if err := v(parsed); err != nil {
				return fmt.Errorf("at %s: %w", path, err)
			}
		}
		return nil
	})
}

// Timestamp checks if the value is a valid timestamp in RFC3339 string format
func Timestamp() Matcher {
	return timeValue(rfc3339Parser)
}

func rfc3339Parser(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// Date checks if the value is a valid date in the format YYYY-MM-DD string
func Date() Matcher {
	return timeValue(dateParser)
}

func dateParser(s string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected YYYY-MM-DD, got %q", s)
	}
	return parsed, nil
}

// TimeWithinDuration checks if the value is a valid time within the specified duration
func TimeWithinDuration(expected time.Time, delta time.Duration) Matcher {
	return timeValue(rfc3339Parser, func(parsed time.Time) error {
		if math.Abs(parsed.Sub(expected).Seconds()) > delta.Seconds() {
			return fmt.Errorf("expected time within %v of %v, got %v", delta, expected, parsed)
		}
		return nil
	})
}

// TimeWithinRange checks if the value is a valid time within the specified range
func TimeWithinRange(startTime, endTime time.Time) Matcher {
	return timeValue(rfc3339Parser, func(parsed time.Time) error {
		if parsed.Before(startTime) || parsed.After(endTime) {
			return fmt.Errorf("expected time between %v and %v, got %v", startTime, endTime, parsed)
		}
		return nil
	})
}

// TimeBefore checks if the value is a valid time before the specified time
func TimeBefore(before time.Time) Matcher {
	return timeValue(rfc3339Parser, func(parsed time.Time) error {
		if !parsed.Before(before) {
			return fmt.Errorf("expected time before %v, got %v", before, parsed)
		}
		return nil
	})
}

// TimeAfter checks if the value is a valid time after the specified time
func TimeAfter(after time.Time) Matcher {
	return timeValue(rfc3339Parser, func(parsed time.Time) error {
		if !parsed.After(after) {
			return fmt.Errorf("expected time after %v, got %v", after, parsed)
		}
		return nil
	})
}

// Number asserts the value is a number
func Number() Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		_, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}
		return nil
	})
}

// NumberWithinDelta asserts the value is a number within a delta of the expected value
func NumberWithinDelta(expected float64, delta float64) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		f64, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}

		if math.Abs(f64-expected) > delta {
			return fmt.Errorf("expected number within %v of %v, got %v", delta, expected, f64)
		}

		return nil
	})
}

// NumberWithinRange asserts the value is a number within a range
func NumberWithinRange(min float64, max float64) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		f64, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}

		if f64 < min || f64 > max {
			return fmt.Errorf("expected number within range %v to %v, got %v", min, max, f64)
		}

		return nil
	})
}

// NumberGreater asserts the value is a number greater than the minimum
func NumberGreater(min float64) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		f64, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}

		if f64 <= min {
			return fmt.Errorf("expected number greater than %v, got %v", min, f64)
		}

		return nil
	})
}

// NumberSmaller asserts the value is a number smaller than the maximum
func NumberSmaller(max float64) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		f64, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}

		if f64 >= max {
			return fmt.Errorf("expected number smaller than %v, got %v", max, f64)
		}

		return nil
	})
}

// Integer asserts the value is an integer
func Integer() Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		f64, ok := value.(float64)
		if !ok {
			return fmt.Errorf("at %s: expected number, got %T", path, value)
		}

		if f64 != math.Trunc(f64) {
			return fmt.Errorf("expected integer, got %v", f64)
		}

		return nil
	})
}

// Positive asserts the value is a positive number
func Positive() Matcher {
	return NumberGreater(0)
}

// Negative asserts the value is a negative number
func Negative() Matcher {
	return NumberSmaller(0)
}

// Object is a function that returns a Matcher that matches a JSON object.
// Extra keys in the actual object are ignored (partial matching).
func Object(expected map[string]any) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		actualMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("at %s: expected object, got %T", path, value)
		}

		for key, expectedVal := range expected {
			actualVal, exists := actualMap[key]
			if !exists {
				return fmt.Errorf("at %s: missing key %q", path, key)
			}

			childPath := fmt.Sprintf("%s.%s", path, key)
			if err := match(expectedVal, childPath, actualVal); err != nil {
				return err
			}
		}

		return nil
	})
}

// StrictObject is a function that returns a Matcher that matches a JSON object.
// Extra keys in the actual object cause a mismatch error.
func StrictObject(expected map[string]any) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		actualMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("at %s: expected object, got %T", path, value)
		}

		for key := range actualMap {
			if _, expectedExists := expected[key]; !expectedExists {
				return fmt.Errorf("at %s: unexpected key %q", path, key)
			}
		}

		for key, expectedVal := range expected {
			actualVal, exists := actualMap[key]
			if !exists {
				return fmt.Errorf("at %s: missing key %q", path, key)
			}

			childPath := fmt.Sprintf("%s.%s", path, key)
			if err := match(expectedVal, childPath, actualVal); err != nil {
				return err
			}
		}

		return nil
	})
}

// Array asserts that the value is an array and matches elements in order.
func Array(elements ...interface{}) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("at %s: expected array, got %T", path, value)
		}

		if len(arr) != len(elements) {
			return fmt.Errorf("at %s: expected array length %d, got %d", path, len(elements), len(arr))
		}

		for i, expected := range elements {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if err := match(expected, childPath, arr[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// UnorderedArray asserts that the value is an array containing the specified elements, in any order.
func UnorderedArray(elements ...interface{}) Matcher {
	return MatcherFunc(func(path string, value interface{}) error {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("at %s: expected array, got %T", path, value)
		}

		if len(arr) != len(elements) {
			return fmt.Errorf("at %s: expected array length %d, got %d", path, len(elements), len(arr))
		}

		// Create a checklist of used indices in the actual array
		used := make([]bool, len(arr))

		// For each expected element, find a match in the actual array that hasn't been used
		for i, expected := range elements {
			found := false
			for j, actual := range arr {
				if used[j] {
					continue
				}

				// Try to match
				// We pass a dummy path because we are just probing
				if err := match(expected, "probe", actual); err == nil {
					used[j] = true
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("at %s: expected element %v (index %d) not found in remaining actual elements", path, expected, i)
			}
		}

		return nil
	})
}
