package bodyguard

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestMatchers(t *testing.T) {
	tests := map[string]struct {
		body     string
		expected interface{}
		wantErr  string
	}{
		// --- Null ---
		"Null Pass": {
			body:     `null`,
			expected: Null(),
			wantErr:  "",
		},
		"Null Fail": {
			body:     `123`,
			expected: Null(),
			wantErr:  "expected null, got 123",
		},

		// --- Bool ---
		"Bool Pass": {
			body:     `true`,
			expected: Bool(),
			wantErr:  "",
		},
		"Bool Fail": {
			body:     `"true"`,
			expected: Bool(),
			wantErr:  "expected boolean, got string",
		},

		// --- String ---
		"String Pass": {
			body:     `"hello"`,
			expected: String(),
			wantErr:  "",
		},
		"String Fail": {
			body:     `123`,
			expected: String(),
			wantErr:  "expected string, got float64",
		},

		// --- UUID ---
		"UUID Pass": {
			body:     `"550e8400-e29b-41d4-a716-446655440000"`,
			expected: UUID(),
			wantErr:  "",
		},
		"UUID Fail": {
			body:     `"not-a-uuid"`,
			expected: UUID(),
			wantErr:  "expected UUID, got \"not-a-uuid\"",
		},

		// --- Email ---
		"Email Pass": {
			body:     `"test@example.com"`,
			expected: Email(),
			wantErr:  "",
		},
		"Email Fail": {
			body:     `"invalid-email"`,
			expected: Email(),
			wantErr:  "expected email, got \"invalid-email\"",
		},

		// --- Regexp ---
		"Regexp Pass": {
			body:     `"abc-123"`,
			expected: Regexp(`^[a-z]{3}-[0-9]{3}$`),
			wantErr:  "",
		},
		"Regexp Fail": {
			body:     `"abcd-123"`,
			expected: Regexp(`^[a-z]{3}-[0-9]{3}$`),
			wantErr:  "expected to match \"^[a-z]{3}-[0-9]{3}$\", got \"abcd-123\"",
		},

		// --- StringLength ---
		"StringLength Pass": {
			body:     `"hello"`,
			expected: StringLength(3, 10),
			wantErr:  "",
		},
		"StringLength Fail": {
			body:     `"hi"`,
			expected: StringLength(3, 10),
			wantErr:  "expected string length between 3 and 10, got 2",
		},

		// --- URL ---
		"URL Pass": {
			body:     `"https://example.com"`,
			expected: URL(),
			wantErr:  "",
		},
		"URL Fail": {
			body:     `"not-a-url"`,
			expected: URL(),
			wantErr:  "expected valid URL, got \"not-a-url\"",
		},

		// --- OneOf ---
		"OneOf Pass": {
			body:     `"apple"`,
			expected: OneOf("apple", "banana", "cherry"),
			wantErr:  "",
		},
		"OneOf Fail": {
			body:     `"pear"`,
			expected: OneOf("apple", "banana", "cherry"),
			wantErr:  "expected one of [apple banana cherry], got \"pear\"",
		},

		// --- Timestamp (formerly RFC3339) ---
		"Timestamp Pass": {
			body:     `"2023-10-27T10:00:00Z"`,
			expected: Timestamp(),
			wantErr:  "",
		},
		"Timestamp Fail": {
			body:     `"2023-10-27"`,
			expected: Timestamp(),
			wantErr:  "cannot parse",
		},

		// --- Date ---
		"Date Pass": {
			body:     `"2023-10-27"`,
			expected: Date(),
			wantErr:  "",
		},
		"Date Fail": {
			body:     `"2023/10/27"`,
			expected: Date(),
			wantErr:  "expected YYYY-MM-DD",
		},

		// --- Time Matchers ---
		"TimeWithinDuration Pass": {
			body:     fmt.Sprintf("%q", time.Now().Format(time.RFC3339)),
			expected: TimeWithinDuration(time.Now(), time.Second),
			wantErr:  "",
		},
		"TimeWithinRange Pass": {
			body:     fmt.Sprintf("%q", time.Now().Format(time.RFC3339)),
			expected: TimeWithinRange(time.Now().Add(-time.Hour), time.Now().Add(time.Hour)),
			wantErr:  "",
		},
		"TimeBefore Pass": {
			body:     fmt.Sprintf("%q", time.Now().Add(-time.Hour).Format(time.RFC3339)),
			expected: TimeBefore(time.Now()),
			wantErr:  "",
		},
		"TimeAfter Pass": {
			body:     fmt.Sprintf("%q", time.Now().Add(time.Hour).Format(time.RFC3339)),
			expected: TimeAfter(time.Now()),
			wantErr:  "",
		},

		// --- Number ---
		"Number Generic Pass": {
			body:     `123.45`,
			expected: Number(),
			wantErr:  "",
		},
		"Number Generic Fail": {
			body:     `"123"`,
			expected: Number(),
			wantErr:  "expected number, got string",
		},
		"Number Value Pass": {
			body:     `123`,
			expected: 123,
			wantErr:  "",
		},
		"Number Value Fail": {
			body:     `123`,
			expected: 456,
			wantErr:  "expected 456 (int), got 123 (float64)",
		},
		"Number Float Pass": {
			body:     `1.23`,
			expected: 1.23,
			wantErr:  "",
		},

		// --- Number Matchers ---
		"NumberWithinDelta Pass": {
			body:     `10.05`,
			expected: NumberWithinDelta(10.0, 0.1),
			wantErr:  "",
		},
		"NumberWithinRange Pass": {
			body:     `50`,
			expected: NumberWithinRange(10, 100),
			wantErr:  "",
		},
		"NumberGreater Pass": {
			body:     `10.1`,
			expected: NumberGreater(10),
			wantErr:  "",
		},
		"NumberSmaller Pass": {
			body:     `9.9`,
			expected: NumberSmaller(10),
			wantErr:  "",
		},
		"NumberGreater Fail": {
			body:     `10`,
			expected: NumberGreater(10),
			wantErr:  "expected number greater than 10, got 10",
		},

		// --- Integer ---
		"Integer Pass": {
			body:     `123`,
			expected: Integer(),
			wantErr:  "",
		},
		"Integer Fail": {
			body:     `123.45`,
			expected: Integer(),
			wantErr:  "expected integer, got 123.45",
		},

		// --- Positive ---
		"Positive Pass": {
			body:     `1`,
			expected: Positive(),
			wantErr:  "",
		},
		"Positive Fail": {
			body:     `0`,
			expected: Positive(),
			wantErr:  "expected number greater than 0, got 0",
		},

		// --- Negative ---
		"Negative Pass": {
			body:     `-1`,
			expected: Negative(),
			wantErr:  "",
		},
		"Negative Fail": {
			body:     `1`,
			expected: Negative(),
			wantErr:  "expected number smaller than 0, got 1",
		},

		// --- Object ---
		"Object Pass": {
			body: `{"a": 1, "b": "s"}`,
			expected: Object(map[string]any{
				"a": 1,
				"b": String(),
			}),
			wantErr: "",
		},
		"Object Partial Pass": {
			body: `{"a": 1, "b": "s", "c": 3}`,
			expected: Object(map[string]any{
				"a": 1,
			}),
			wantErr: "",
		},
		"Object Missing Key": {
			body: `{"a": 1}`,
			expected: Object(map[string]any{
				"b": 2,
			}),
			wantErr: "missing key \"b\"",
		},
		"Object Type Mismatch": {
			body: `{"a": 1}`,
			expected: Object(map[string]any{
				"a": String(),
			}),
			wantErr: "expected string, got float64",
		},

		// --- StrictObject ---
		"StrictObject Pass": {
			body: `{"a": 1, "b": 2}`,
			expected: StrictObject(map[string]any{
				"a": 1,
				"b": 2,
			}),
			wantErr: "",
		},
		"StrictObject Extra Key": {
			body: `{"a": 1, "b": 2, "c": 3}`,
			expected: StrictObject(map[string]any{
				"a": 1,
				"b": 2,
			}),
			wantErr: "unexpected key \"c\"",
		},
		"StrictObject Missing Key": {
			body: `{"a": 1}`,
			expected: StrictObject(map[string]any{
				"a": 1,
				"b": 2,
			}),
			wantErr: "missing key \"b\"",
		},

		// --- Array ---
		"Array Pass": {
			body:     `[1, "two", true]`,
			expected: Array(1, "two", Bool()),
			wantErr:  "",
		},
		"Array Length Mismatch": {
			body:     `[1, 2]`,
			expected: Array(1, 2, 3),
			wantErr:  "expected array length 3, got 2",
		},
		"Array Value Mismatch": {
			body:     `[1, 2, 3]`,
			expected: Array(1, 4, 3),
			wantErr:  "expected 4 (int), got 2 (float64)",
		},

		// --- UnorderedArray ---
		"UnorderedArray Pass": {
			body:     `[3, 1, 2]`,
			expected: UnorderedArray(1, 2, 3),
			wantErr:  "",
		},
		"UnorderedArray with Matchers Pass": {
			body:     `[3, "two", 1]`,
			expected: UnorderedArray(1, String(), 3),
			wantErr:  "",
		},
		"UnorderedArray Fail": {
			body:     `[1, 2, 3]`,
			expected: UnorderedArray(1, 2, 4),
			wantErr:  "element 4 (index 2) not found",
		},

		// --- StringWithFormat ---
		"StringWithFormat Pass": {
			body: `"FOO"`,
			expected: StringWithFormat(func(s string) error {
				if s != "FOO" {
					return fmt.Errorf("expected FOO")
				}
				return nil
			}),
			wantErr: "",
		},
		"StringWithFormat Fail": {
			body: `"BAR"`,
			expected: StringWithFormat(func(s string) error {
				if s != "FOO" {
					return fmt.Errorf("expected FOO")
				}
				return nil
			}),
			wantErr: "expected FOO",
		},

		// --- Literal Matches ---
		"Literal Int Pass": {
			body:     `123`,
			expected: 123,
			wantErr:  "",
		},
		"Literal String Pass": {
			body:     `"foo"`,
			expected: "foo",
			wantErr:  "",
		},
		"Literal Bool Pass": {
			body:     `true`,
			expected: true,
			wantErr:  "",
		},
		"Literal Float Pass": {
			body:     `1.23`,
			expected: 1.23,
			wantErr:  "",
		},
		"Literal Mismatch": {
			body:     `"foo"`,
			expected: "bar",
			wantErr:  "expected bar (string), got foo (string)",
		},
		"Literal Type Mismatch": {
			body:     `true`,
			expected: "true",
			wantErr:  "expected true (string), got true (bool)",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := isMatch(tt.body, tt.expected)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.wantErr)
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
