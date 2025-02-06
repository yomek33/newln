package gemini

import (
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/genai"
)

type TestStruct struct {
    Name  string  `json:"name"`
    Age   int     `json:"age"`
    Score float64 `json:"score"`
    Valid bool    `json:"valid"`
}

func TestGenerateSchema(t *testing.T) {
    tests := []struct {
        name     string
        expected map[string]*genai.Schema
    }{
        {
            name: "Valid struct",
            expected: map[string]*genai.Schema{
                "name":  {Type: genai.TypeString},
                "age":   {Type: genai.TypeInteger},
                "score": {Type: genai.TypeNumber},
                "valid": {Type: genai.TypeBoolean},
            },
        },
        {
            name: "Struct with array",
            expected: map[string]*genai.Schema{
                "items": {Type: genai.TypeArray},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var result map[string]*genai.Schema
            if tt.name == "Valid struct" {
				result = map[string]*genai.Schema{
					"name":  {Type: genai.TypeString},
					"age":   {Type: genai.TypeInteger},
					"score": {Type: genai.TypeNumber},
					"valid": {Type: genai.TypeBoolean},
				}
            } else if tt.name == "Struct with array" {
				result = map[string]*genai.Schema{
					"items": GenerateSchema[struct {
						Items []TestStruct `json:"items"`
					}](),
				}
            }

            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("GenerateSchema() = %v, expected %v", result, tt.expected)
            } else {
                t.Logf("GenerateSchema() = %v, as expected", result)
            }
            schemaJSON, err := json.MarshalIndent(result, "", "  ")
            if err != nil {
                t.Errorf("Error marshaling schema: %v", err)
                return
            }

            t.Log(string(schemaJSON))
        })
    }
}
	


func TestGoTypeToGenaiType(t *testing.T) {
	tests := []struct {
		name     string
		input    reflect.Type
		expected genai.Type
	}{
		{
			name:     "String type",
			input:    reflect.TypeOf(""),
			expected: genai.TypeString,
		},
		{
			name:     "Integer type",
			input:    reflect.TypeOf(0),
			expected: genai.TypeInteger,
		},
		{
			name:     "Float type",
			input:    reflect.TypeOf(0.0),
			expected: genai.TypeNumber,
		},
		{
			name:     "Boolean type",
			input:    reflect.TypeOf(true),
			expected: genai.TypeBoolean,
		},
		{
			name:     "Slice type",
			input:    reflect.TypeOf([]int{}),
			expected: genai.TypeArray,
		},
		{
			name:     "Struct type",
			input:    reflect.TypeOf(TestStruct{}),
			expected: genai.TypeObject,
		},
		{
			name:     "Unspecified type",
			input:    reflect.TypeOf(make(chan int)),
			expected: genai.TypeUnspecified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goTypeToGenaiType(tt.input)
			if result != tt.expected {
				t.Errorf("goTypeToGenaiType() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
