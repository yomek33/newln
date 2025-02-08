package vertex

import (
	"encoding/json"
	"fmt"
	"reflect"

	"cloud.google.com/go/vertexai/genai"
)

// ジェネリックな JSON デコード関数
func DecodeJsonContent[T any](data json.RawMessage) (T, error) {
	var output T
	if err := json.Unmarshal(data, &output); err != nil {
		return output, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return output, nil
}

func GenerateSchema[T any]() *genai.Schema {
	var t T
	return generateSchema(reflect.TypeOf(t))
}

// 型 (`reflect.Type`) を直接渡す
func generateSchema(t reflect.Type) *genai.Schema {
	// ポインタの場合はデリファレンス
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &genai.Schema{Type: genai.TypeString}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &genai.Schema{Type: genai.TypeInteger}
	case reflect.Float32, reflect.Float64:
		return &genai.Schema{Type: genai.TypeNumber}
	case reflect.Bool:
		return &genai.Schema{Type: genai.TypeBoolean}
	case reflect.Slice, reflect.Array:
		return &genai.Schema{
			Type:  genai.TypeArray,
			Items: generateSchema(t.Elem()), //再帰
		}
	case reflect.Struct:
		properties := make(map[string]*genai.Schema)
		requiredFields := getRequiredFields(t)

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			properties[jsonTag] = generateSchema(field.Type)
		}

		return &genai.Schema{
			Type:       genai.TypeObject,
			Properties: properties,
			Required:   requiredFields,
		}
	default:
		return &genai.Schema{Type: genai.TypeUnspecified}
	}
}

// 必須フィールドを取得
func getRequiredFields(t reflect.Type) []string {
	var required []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			required = append(required, jsonTag)
		}
	}
	return required
}

// goTypeToGenaiType: Go の型を genai の型に変換
func goTypeToGenaiType(t reflect.Type) genai.Type {
	switch t.Kind() {
	case reflect.String:
		return genai.TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return genai.TypeInteger
	case reflect.Float32, reflect.Float64:
		return genai.TypeNumber
	case reflect.Bool:
		return genai.TypeBoolean
	case reflect.Slice, reflect.Array:
		return genai.TypeArray
	case reflect.Struct:
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}
