package bodyguard_test

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

func TestExample_APIResponse(t *testing.T) {
	jsonPayload := `{
		"meta": {
			"page": 1,
			"total_pages": 5,
			"total_items": 42
		},
		"data": [
			{
				"id": 1,
				"name": "Widget A",
				"price": 19.99
			},
			{
				"id": 2,
				"name": "Widget B",
				"price": 25.50
			}
		]
	}`

	bodyguard.Assert(t, bodyguard.Object(map[string]any{
		"meta": bodyguard.Object(map[string]any{
			"page":        1,
			"total_pages": bodyguard.Number(),
			"total_items": bodyguard.NumberGreater(0),
		}),
		"data": bodyguard.Array(
			bodyguard.Object(map[string]any{
				"id":    bodyguard.Number(),
				"name":  bodyguard.String(),
				"price": 19.99,
			}),
			bodyguard.Object(map[string]any{
				"id":    bodyguard.Number(),
				"name":  bodyguard.String(),
				"price": bodyguard.Number(),
			}),
		),
	}), jsonPayload)
}

func TestExample_StrictValidation(t *testing.T) {
	// StrictObject ensures no extra fields are present in the Assert
	jsonPayload := `{
		"id": 1,
		"name": "Strict Item"
	}`

	bodyguard.Assert(t, bodyguard.StrictObject(map[string]any{
		"id":   1,
		"name": "Strict Item",
	}), jsonPayload)
}

func TestExample_UnorderedList(t *testing.T) {
	// UnorderedArray allows elements to appear in any order
	jsonPayload := `["apple", "banana", "cherry"]`

	bodyguard.Assert(t, bodyguard.UnorderedArray(
		"cherry",
		"apple",
		"banana",
	), jsonPayload)
}
